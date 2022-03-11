# SPDX-License-Identifier: GPL-3.0-or-later

"""External dependencies loader."""

import contextlib
import importlib
from pathlib import Path
import sys
import logging
from types import ModuleType
from typing import Iterator

_my_dir = Path(__file__).parent
_log = logging.getLogger(__name__)


def load_wheel(module_name: str, fname_prefix: str = "") -> ModuleType:
    """Loads a wheel from 'fname_prefix*.whl', unless the named module can be imported.

    This allows us to use system-installed packages before falling back to the shipped wheels.
    This is useful for development, less so for deployment.

    If `fname_prefix` is the empty string, it will use the module name.
    """

    if not fname_prefix:
        fname_prefix = _fname_prefix_from_module_name(module_name)

    try:
        module = importlib.import_module(module_name)
    except ImportError as ex:
        _log.debug("Unable to import %s directly, will try wheel: %s", module_name, ex)
    else:
        _log.debug(
            "Was able to load %s from %s, no need to load wheel %s",
            module_name,
            module.__file__,
            fname_prefix,
        )
        assert isinstance(module, ModuleType)
        return module

    wheel = _wheel_filename(fname_prefix)

    # Load the module from the wheel file. Keep a backup of sys.path so that it
    # can be restored later. This should ensure that future import statements
    # cannot find this wheel file, increasing the separation of dependencies of
    # this add-on from other add-ons.
    with _sys_path_mod_backup(wheel):
        try:
            module = importlib.import_module(module_name)
        except ImportError as ex:
            raise ImportError(
                "Unable to load %r from %s: %s" % (module_name, wheel, ex)
            ) from None

    _log.debug("Loaded %s from %s", module_name, module.__file__)
    assert isinstance(module, ModuleType)
    return module


def load_wheel_global(module_name: str, fname_prefix: str = "") -> ModuleType:
    """Loads a wheel from 'fname_prefix*.whl', unless the named module can be imported.

    This allows us to use system-installed packages before falling back to the shipped wheels.
    This is useful for development, less so for deployment.

    If `fname_prefix` is the empty string, it will use the first package from `module_name`.
    In other words, `module_name="pkg.subpkg"` will result in `fname_prefix="pkg"`.
    """

    if not fname_prefix:
        fname_prefix = _fname_prefix_from_module_name(module_name)

    try:
        module = importlib.import_module(module_name)
    except ImportError as ex:
        _log.debug("Unable to import %s directly, will try wheel: %s", module_name, ex)
    else:
        _log.debug(
            "Was able to load %s from %s, no need to load wheel %s",
            module_name,
            module.__file__,
            fname_prefix,
        )
        return module

    wheel = _wheel_filename(fname_prefix)

    wheel_filepath = str(wheel)
    if wheel_filepath not in sys.path:
        sys.path.insert(0, wheel_filepath)

    try:
        module = importlib.import_module(module_name)
    except ImportError as ex:
        raise ImportError(
            "Unable to load %r from %s: %s" % (module_name, wheel, ex)
        ) from None

    _log.debug("Globally loaded %s from %s", module_name, module.__file__)
    return module


@contextlib.contextmanager
def _sys_path_mod_backup(wheel_file: Path) -> Iterator[None]:
    """Temporarily inserts a wheel onto sys.path.

    When the context exits, it restores sys.path and sys.modules, so that
    anything that was imported within the context remains unimportable by other
    modules.
    """
    old_syspath = sys.path[:]
    old_sysmod = sys.modules.copy()

    try:
        sys.path.insert(0, str(wheel_file))
        yield
    finally:
        # Restore without assigning a new list instance. That way references
        # held by other code will stay valid.
        sys.path[:] = old_syspath
        sys.modules.clear()
        sys.modules.update(old_sysmod)


def _wheel_filename(fname_prefix: str) -> Path:
    path_pattern = "%s*.whl" % fname_prefix
    wheels: list[Path] = list(_my_dir.glob(path_pattern))
    if not wheels:
        raise RuntimeError("Unable to find wheel at %r" % path_pattern)

    # If there are multiple wheels that match, load the last-modified one.
    # Alphabetical sorting isn't going to cut it since BAT 1.10 was released.
    def modtime(filepath: Path) -> float:
        return filepath.stat().st_mtime

    wheels.sort(key=modtime)
    return wheels[-1]


def _fname_prefix_from_module_name(module_name: str) -> str:
    return module_name.split(".", 1)[0]


if __name__ == "__main__":
    logging.basicConfig(level=logging.DEBUG)
    wheel = _wheel_filename("python_dateutil")
    print(f"Wheel: {wheel}")
    module = load_wheel("dateutil", "python_dateutil")
    print(f"module: {module}")
