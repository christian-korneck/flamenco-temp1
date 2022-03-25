# SPDX-License-Identifier: GPL-3.0-or-later
"""BAT interface for sending files to the Manager via the Shaman API."""

import logging
import random
from collections import deque
from pathlib import Path, PurePath, PurePosixPath
from typing import TYPE_CHECKING, Optional, Any

from .. import wheels
from . import cache

bat_pack = wheels.load_wheel("blender_asset_tracer.pack")
bat_transfer = wheels.load_wheel("blender_asset_tracer.pack.transfer")
bat_bpathlib = wheels.load_wheel("blender_asset_tracer.bpathlib")


if TYPE_CHECKING:
    from ..manager import ApiClient as _ApiClient

    from ..manager.models import (
        ShamanRequirementsRequest as _ShamanRequirementsRequest,
        ShamanRequirementsResponse as _ShamanRequirementsResponse,
        ShamanFileSpec as _ShamanFileSpec,
    )
else:
    _ApiClient = object
    _ShamanRequirementsRequest = object
    _ShamanRequirementsResponse = object
    _ShamanFileSpec = object

log = logging.getLogger(__name__)

MAX_DEFERRED_PATHS = 8
MAX_FAILED_PATHS = 8


# Mypy doesn't understand that bat_pack.Packer exists.
class Packer(bat_pack.Packer):  # type: ignore
    """Creates BAT Packs on a Shaman server."""

    def __init__(
        self,
        blendfile: Path,
        project_root: Path,
        target: str,
        *,
        api_client: _ApiClient,
        checkout_path: str,
        **kwargs: dict[Any, Any],
    ) -> None:
        """Constructor

        :param target: mock target root directory to construct project-relative paths.
        """
        super().__init__(blendfile, project_root, target, **kwargs)
        self.checkout_path = checkout_path
        self.api_client = api_client

    # Mypy doesn't understand that bat_transfer.FileTransferer exists.
    def _create_file_transferer(self) -> bat_transfer.FileTransferer:  # type: ignore
        return Transferrer(self.api_client, self.project, self.checkout_path)

    def _make_target_path(self, target: str) -> PurePath:
        return PurePosixPath("/")

    @property
    def output_path(self) -> PurePath:
        """The path of the packed blend file in the target directory."""
        assert self._output_path is not None

        checkout_location = PurePosixPath(self.checkout_path)
        rel_output = self._output_path.relative_to(self._target_path)
        out_path: PurePath = checkout_location / rel_output
        return out_path

    def execute(self):
        try:
            super().execute()
        except Exception as ex:
            log.exception("Error communicating with Shaman")
            self.abort(str(ex))
            self._check_aborted()


class Transferrer(bat_transfer.FileTransferer):  # type: ignore
    """Sends files to a Shaman server."""

    class AbortUpload(Exception):
        """Raised from the upload callback to abort an upload."""

    def __init__(
        self,
        api_client: _ApiClient,
        local_project_root: Path,
        checkout_path: str,
    ) -> None:
        super().__init__()
        from ..manager.apis import ShamanApi

        self.shaman_api = ShamanApi(api_client)

        self.project_root = local_project_root
        self.checkout_path = checkout_path
        self.log = log.getChild(self.__class__.__name__)
        self.uploaded_files = 0
        self.uploaded_bytes = 0

        # Mapping from the relative path (as used in Shaman requests) to the
        # absolute path where we can find the local file. This is typically just
        # the same as the relative path (relative to the project root), but can
        # also point to a temporary file when it had to be rewritten.
        self._rel_to_local_path: dict[str, Path] = {}

        # Temporary files that should be deleted before stopping.
        self._delete_when_done: list[Path] = []

    # noinspection PyBroadException
    def run(self) -> None:
        try:
            self._run()
        except Exception as ex:
            # We have to catch exceptions in a broad way, as this is running in
            # a separate thread, and exceptions won't otherwise be seen.
            self.log.exception("Error transferring files to Shaman")
            self.error_set("Unexpected exception transferring files to Shaman: %s" % ex)
        finally:
            # Delete the files that were supposed to be moved.
            for src in self._delete_when_done:
                self.delete_file(src)
            self._delete_when_done.clear()

    def _run(self) -> None:
        self.uploaded_files = 0
        self.uploaded_bytes = 0

        # Construct the Shaman Checkout Definition file.
        shaman_file_specs = self._create_checkout_definition()
        if not shaman_file_specs:
            # An error has already been logged.
            return

        failed_files = self._upload_missing_files(shaman_file_specs)
        if failed_files:
            self.log.error("Aborting upload due to too many failures")
            self.error_set("Giving up after multiple attempts to upload the files")
            return

        self.log.info("All files uploaded succesfully")
        self._request_checkout(shaman_file_specs)

    def _upload_missing_files(
        self, shaman_file_specs: _ShamanRequirementsRequest
    ) -> set[_ShamanFileSpec]:
        self.log.info("Feeding %d files to the Shaman", len(shaman_file_specs.files))
        if self.log.isEnabledFor(logging.INFO):
            for spec in shaman_file_specs.files:
                self.log.info("   - %s", spec.path)

        # Try to upload all the files.
        failed_files: set[_ShamanFileSpec] = set()
        max_tries = 50
        for try_index in range(max_tries):
            # Send the file to the Shaman and see what we still need to send there.
            to_upload = self._send_checkout_def_to_shaman(shaman_file_specs)
            if to_upload is None:
                # An error has already been logged.
                return failed_files

            if not to_upload:
                break

            # Send the files that still need to be sent.
            self.log.info("Upload attempt %d", try_index + 1)
            failed_files = self._upload_files(to_upload)
            if not failed_files:
                break

            # Having failed paths at this point is expected when multiple
            # clients are sending the same files. Instead of retrying on a
            # file-by-file basis, we just re-send the checkout definition
            # file to the Shaman and obtain a new list of files to upload.
        return failed_files

    def _create_checkout_definition(self) -> Optional[_ShamanRequirementsRequest]:
        """Create the checkout definition file for this BAT pack.

        :returns: the checkout definition.

        If there was an error and file transfer was aborted, the checkout
        definition will be empty.
        """

        from ..manager.models import (
            ShamanRequirementsRequest,
            ShamanFileSpec,
        )

        filespecs: list[ShamanFileSpec] = []

        for src, dst, act in self.iter_queue():
            try:
                checksum = cache.compute_cached_checksum(src)
                filesize = src.stat().st_size
                relpath = str(bat_bpathlib.strip_root(dst))

                filespec = ShamanFileSpec(
                    sha=checksum,
                    size=filesize,
                    path=relpath,
                )
                filespecs.append(filespec)
                self._rel_to_local_path[relpath] = src

                if act == bat_transfer.Action.MOVE:
                    self._delete_when_done.append(src)
            except Exception:
                # We have to catch exceptions in a broad way, as this is running in
                # a separate thread, and exceptions won't otherwise be seen.
                msg = "Error transferring %s to %s" % (src, dst)
                self.log.exception(msg)
                # Put the files to copy back into the queue, and abort. This allows
                # the main thread to inspect the queue and see which files were not
                # copied. The one we just failed (due to this exception) should also
                # be reported there.
                self.queue.put((src, dst, act))
                self.error_set(msg)
                return None

        cache.cleanup_cache()
        specs: ShamanRequirementsRequest = ShamanRequirementsRequest(files=filespecs)
        return specs

    def _send_checkout_def_to_shaman(
        self,
        requirements: _ShamanRequirementsRequest,
    ) -> Optional[deque[_ShamanFileSpec]]:
        """Send the checkout definition file to the Shaman.

        :return: An iterable of file specs that still need to be uploaded, or
            None if there was an error.
        """
        from ..manager.exceptions import ApiException

        try:
            resp = self.shaman_api.shaman_checkout_requirements(requirements)
        except ApiException as ex:
            # TODO: the body should be JSON of a predefined type, parse it to get the actual message.
            msg = "Error from Shaman, code %d: %s" % (ex.status, ex.body)
            self.log.error(msg)
            self.error_set(msg)
            return None
        assert isinstance(resp, _ShamanRequirementsResponse)

        to_upload: deque[_ShamanFileSpec] = deque()
        for file_spec in resp.files:
            print(file_spec)
            if file_spec.path not in self._rel_to_local_path:
                msg = (
                    "Shaman requested path we did not intend to upload: %r" % file_spec
                )
                self.log.error(msg)
                self.error_set(msg)
                return None

            self.log.debug("   %s: %s", file_spec.status.value, file_spec.path)
            match file_spec.status.value:
                case "unknown":
                    to_upload.appendleft(file_spec)
                case "uploading":
                    to_upload.append(file_spec)
                case _:
                    msg = "Unknown status in response from Shaman: %r" % file_spec
                    self.log.error(msg)
                    self.error_set(msg)
                    return None
        return to_upload

    def _upload_files(self, to_upload: deque[_ShamanFileSpec]) -> set[_ShamanFileSpec]:
        """Actually upload the files to Shaman.

        Returns the set of files that we did not upload.
        """
        if not to_upload:
            self.log.info("All files are at the Shaman already")
            self.report_transferred(0)
            return set()

        from ..manager.exceptions import ApiException

        failed_specs: set[_ShamanFileSpec] = set()
        deferred_specs: set[_ShamanFileSpec] = set()

        def defer(filespec: _ShamanFileSpec) -> None:
            nonlocal to_upload

            self.log.info(
                "   %s deferred (already being uploaded by someone else)", filespec.path
            )
            deferred_specs.add(filespec)

            # Instead of deferring this one file, randomize the files to upload.
            # This prevents multiple deferrals when someone else is uploading
            # files from the same project (because it probably happens alphabetically).
            all_files = list(to_upload)
            random.shuffle(all_files)
            to_upload = deque(all_files)

        self.log.info(
            "Going to upload %d of %d files", len(to_upload), len(self._spec_to_paths)
        )
        while to_upload:
            # After too many failures, just retry to get a fresh set of files to upload.
            if len(failed_specs) > MAX_FAILED_PATHS:
                self.log.info("Too many failures, going to abort this iteration")
                failed_specs.update(to_upload)
                return failed_specs

            file_spec = to_upload.popleft()
            self.log.info("   %s", file_spec.path)

            # Pre-flight check. The generated API code will load the entire file into
            # memory before sending it to the Shaman. It's faster to do a check at
            # Shaman first, to see if we need uploading at all.
            check_resp = self.shaman_api.shaman_file_store_check(
                checksum=file_spec.sha,
                filesize=file_spec.size,
            )
            if check_resp.status.value == "stored":
                self.log.info("  %s: skipping, already on server", file_spec.path)
                continue

            # Let the Shaman know whether we can defer uploading this file or not.
            can_defer = (
                len(deferred_specs) < MAX_DEFERRED_PATHS
                and file_spec not in deferred_specs
                and len(to_upload)
            )

            local_filepath = self._rel_to_local_path[file_spec.path]
            try:
                with local_filepath.open("rb") as file_reader:
                    self.shaman_api.shaman_file_store(
                        checksum=file_spec.sha,
                        filesize=file_spec.size,
                        body=file_reader,
                        x_shaman_can_defer_upload=can_defer,
                        x_shaman_original_filename=file_spec.path,
                    )
            except ApiException as ex:
                match ex.status:
                    case 425:  # Too Early, i.e. defer uploading this file.
                        self.log.info(
                            "  %s: someone else is uploading this file, deferring",
                            file_spec.path,
                        )
                        defer(file_spec)
                        continue
                    case 417:  # Expectation Failed; mismatch of checksum or file size.
                        msg = "Error from Shaman uploading %s, code %d: %s" % (
                            file_spec.path,
                            ex.status,
                            ex.body,
                        )
                    case _:  # Unknown error
                        msg = "API exception\nHeaders: %s\nBody: %s\n" % (
                            ex.headers,
                            ex.body,
                        )

                self.log.error(msg)
                self.error_set(msg)
                failed_specs.add(file_spec)
                return failed_specs

            failed_specs.discard(file_spec)
            self.uploaded_files += 1
            file_size = local_filepath.stat().st_size
            self.uploaded_bytes += file_size
            self.report_transferred(file_size)

        if failed_specs:
            self.log.info(
                "Uploaded %d bytes in %d files so far",
                self.uploaded_bytes,
                self.uploaded_files,
            )
            return failed_specs

        self.log.info(
            "Done uploading %d bytes in %d files",
            self.uploaded_bytes,
            self.uploaded_files,
        )
        return set()

    def report_transferred(self, bytes_transferred: int) -> None:
        if self._abort.is_set():
            self.log.warning("Interrupting ongoing upload")
            raise self.AbortUpload("interrupting ongoing upload")
        super().report_transferred(bytes_transferred)

    def _request_checkout(self, shaman_file_specs: _ShamanRequirementsRequest) -> None:
        """Ask the Shaman to create a checkout of this BAT pack."""

        if not self.checkout_path:
            self.log.warning("NOT requesting checkout at Shaman")
            return

        from ..manager.models import ShamanCheckout
        from ..manager.exceptions import ApiException

        self.log.info(
            "Requesting checkout at Shaman for checkout_path=%r", self.checkout_path
        )

        checkoutRequest = ShamanCheckout(
            files=shaman_file_specs.files,
            checkout_path=self.checkout_path,
        )

        try:
            self.shaman_api.shaman_checkout(checkoutRequest)
        except ApiException as ex:
            match ex.status:
                case 424:  # Files were missing
                    msg = "We did not upload some files, checkout aborted"
                case 409:  # Checkout already exists
                    msg = (
                        "There is already an existing checkout at %s"
                        % self.checkout_path
                    )
                case _:  # Unknown error
                    msg = "API exception\nHeaders: %s\nBody: %s\n" % (
                        ex.headers,
                        ex.body,
                    )
            self.log.error(msg)
            self.error_set(msg)
            return

        self.log.info("Shaman created checkout at %s", self.checkout_path)
