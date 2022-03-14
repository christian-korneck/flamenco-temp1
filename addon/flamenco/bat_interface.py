"""BAT packing interface for Flamenco."""

from dataclasses import dataclass
from pathlib import Path
from typing import Optional, Any
import logging
import queue
import threading
import typing

from . import wheels

pack = wheels.load_wheel("blender_asset_tracer.pack")
progress = wheels.load_wheel("blender_asset_tracer.pack.progress")
transfer = wheels.load_wheel("blender_asset_tracer.pack.transfer")
shaman = wheels.load_wheel("blender_asset_tracer.pack.shaman")

log = logging.getLogger(__name__)

# # For using in other parts of the add-on, so only this file imports BAT.
# Aborted = pack.Aborted
# FileTransferError = transfer.FileTransferError
# parse_shaman_endpoint = shaman.parse_endpoint


class Message:
    """Superclass for message objects queued by the BatProgress class."""


@dataclass
class MsgSetWMAttribute(Message):
    """Set a WindowManager attribute to a value."""

    attribute_name: str
    value: object


@dataclass
class MsgException(Message):
    """Report an exception."""

    ex: BaseException


@dataclass
class MsgProgress(Message):
    """Report packing progress."""

    percentage: int  # 0 - 100


@dataclass
class MsgDone(Message):
    output_path: Path
    missing_files: list[Path]


# MyPy doesn't understand the way BAT subpackages are imported.
class BatProgress(progress.Callback):  # type: ignore
    """Report progress of BAT Packing to the given queue."""

    def __init__(self, queue: queue.SimpleQueue[Message]) -> None:
        super().__init__()
        self.queue = queue

    def _set_attr(self, attr: str, value: Any) -> None:
        msg = MsgSetWMAttribute(attr, value)
        self.queue.put(msg)

    def _txt(self, msg: str) -> None:
        """Set a text in a thread-safe way."""
        self._set_attr("flamenco_bat_status_txt", msg)

    def _status(self, status: str) -> None:
        """Set the flamenco_bat_status property in a thread-safe way."""
        self._set_attr("flamenco_bat_status", status)

    def _progress(self, progress: int) -> None:
        """Set the flamenco_bat_progress property in a thread-safe way."""
        self._set_attr("flamenco_bat_progress", progress)
        msg = MsgProgress(percentage=progress)
        self.queue.put(msg)

    def pack_start(self) -> None:
        self._txt("Starting BAT Pack operation")

    def pack_done(
        self, output_blendfile: Path, missing_files: typing.Set[Path]
    ) -> None:
        if missing_files:
            self._txt("There were %d missing files" % len(missing_files))
        else:
            self._txt("Pack of %s done" % output_blendfile.name)

    def pack_aborted(self, reason: str) -> None:
        self._txt("Aborted: %s" % reason)
        self._status("ABORTED")

    def trace_blendfile(self, filename: Path) -> None:
        """Called for every blendfile opened when tracing dependencies."""
        self._txt("Inspecting %s" % filename.name)

    def trace_asset(self, filename: Path) -> None:
        if filename.stem == ".blend":
            return
        self._txt("Found asset %s" % filename.name)

    def rewrite_blendfile(self, orig_filename: Path) -> None:
        self._txt("Rewriting %s" % orig_filename.name)

    def transfer_file(self, src: Path, dst: Path) -> None:
        self._txt("Transferring %s" % src.name)

    def transfer_file_skipped(self, src: Path, dst: Path) -> None:
        self._txt("Skipped %s" % src.name)

    def transfer_progress(self, total_bytes: int, transferred_bytes: int) -> None:
        self._progress(round(100 * transferred_bytes / total_bytes))

    def missing_file(self, filename: Path) -> None:
        # TODO(Sybren): report missing files in a nice way
        pass


# class ShamanPacker(shaman.ShamanPacker):
#     """Packer with support for getting an auth token from Flamenco Server."""

#     def __init__(
#         self,
#         bfile: Path,
#         project: Path,
#         target: str,
#         endpoint: str,
#         checkout_id: str,
#         *,
#         manager_id: str,
#         **kwargs
#     ) -> None:
#         self.manager_id = manager_id
#         super().__init__(bfile, project, target, endpoint, checkout_id, **kwargs)

#     def _get_auth_token(self) -> str:
#         """get a token from Flamenco Server"""

#         from ..blender import PILLAR_SERVER_URL
#         from ..pillar import blender_id_subclient, uncached_session, SUBCLIENT_ID

#         url = urllib.parse.urljoin(
#             PILLAR_SERVER_URL, "flamenco/jwt/generate-token/%s" % self.manager_id
#         )
#         auth_token = blender_id_subclient()["token"]

#         resp = uncached_session.get(url, auth=(auth_token, SUBCLIENT_ID))
#         resp.raise_for_status()
#         return resp.text


class PackThread(threading.Thread):
    queue: queue.SimpleQueue[Message]

    # MyPy doesn't understand the way BAT subpackages are imported.
    def __init__(self, packer: pack.Packer) -> None:  # type: ignore
        # Quitting Blender should abort the transfer (instead of hanging until
        # the transfer is done), hence daemon=True.
        super().__init__(daemon=True, name="PackThread")

        self.queue = queue.SimpleQueue()

        self.packer = packer
        self.packer.progress_cb = BatProgress(queue=self.queue)

    def run(self) -> None:
        global _running_packthread

        try:
            self._run()
        except BaseException as ex:
            log.error("Error packing with BAT: %s", ex)
            self.queue.put(MsgException(ex=ex))
        finally:
            with _packer_lock:
                _running_packthread = None

    def _run(self) -> None:
        with self.packer:
            log.debug("awaiting strategise")
            self._set_bat_status("INVESTIGATING")
            self.packer.strategise()

            log.debug("awaiting execute")
            self._set_bat_status("TRANSFERRING")
            self.packer.execute()

            log.debug("done")
            self._set_bat_status("DONE")

            msg = MsgDone(self.packer.output_path, self.packer.missing_files)
            self.queue.put(msg)

    def _set_bat_status(self, status: str) -> None:
        self.queue.put(MsgSetWMAttribute("flamenco_bat_status", status))

    def poll(self, timeout: Optional[int] = None) -> Optional[Message]:
        """Poll the queue, return the first message or None if there is none.

        :param timeout: Max time to wait for a message to appear on the queue.
            If None, will not wait and just return None immediately (if there is
            no queued message).
        """
        try:
            return self.queue.get(block=timeout is not None, timeout=timeout)
        except queue.Empty:
            return None

    def abort(self) -> None:
        """Abort the running pack operation."""
        self.packer.abort()


_running_packthread: typing.Optional[PackThread] = None
_packer_lock = threading.RLock()


def copy(  # type: ignore
    base_blendfile: Path,
    project: Path,
    target: str,
    exclusion_filter: str,
    *,
    relative_only: bool,
    packer_class=pack.Packer,  # type: ignore
    **packer_args: dict[Any, Any],
) -> PackThread:
    """Use BAT to copy the given file and dependencies to the target location.

    Runs BAT in a separate thread, and returns early. Use poll() to get updates
    & the final result.
    """
    global _running_packthread

    with _packer_lock:
        if _running_packthread is not None:
            raise RuntimeError("other packing operation already in progress")

    packer = packer_class(
        base_blendfile,
        project,
        target,
        compress=True,
        relative_only=relative_only,
        **packer_args,
    )
    if exclusion_filter:
        filter_parts = exclusion_filter.strip().split(" ")
        packer.exclude(*filter_parts)

    packthread = PackThread(packer=packer)
    with _packer_lock:
        _running_packthread = packthread

    packthread.start()
    return packthread


def abort() -> None:
    """Abort a running copy() call.

    No-op when there is no running copy(). Can be called from any thread.
    """

    with _packer_lock:
        if _running_packthread is None:
            log.debug("No running packer, ignoring call to abort()")
            return
        log.info("Aborting running packer")
        _running_packthread.abort()


def is_packing() -> bool:
    """Returns whether a BAT packing operation is running."""

    with _packer_lock:
        return _running_packthread is not None
