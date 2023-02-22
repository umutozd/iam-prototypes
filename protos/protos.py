#!/usr/bin/env python3
import os
import subprocess
import sys
from platform import machine
import json
import glob as _glob
import atexit
import shutil
from typing import Any

# registered calls that will be called with at program exit
_registered = []


def cwd():
    """Returns the directory of the script currently running."""
    return os.path.dirname(os.path.realpath(sys.argv[0]))


def root_dir():
    """Returns the path to the root of the project."""
    c = cwd()
    while c not in ('', '/'):
        p1 = os.path.join(c, 'protos', 'protos.py')
        p2 = os.path.join(c, 'apis', 'protos', 'protos.py')
        if os.path.exists(p1) or os.path.exists(p2):
            return c
        else:
            c = os.path.dirname(c)
    # happens only if cwd is not in the project
    raise Exception('unable to determine root_dir')


def import_path(extra_imports=[]):
    """Returns the default imports and extra ones if specified."""
    root = root_dir()
    if isinstance(extra_imports, str):
        extra_imports = [extra_imports]
    i = ':'.join([
        cwd(),
        root,
        os.path.join(root, 'vendor'),
        os.path.join(root, 'protos'),
        os.path.join(root, 'vendor', 'github.com', 'gogo', 'protobuf'),
        # os.path.join(
        #     root, 'vendor', 'github.com', 'otsimo', 'otsimopb')
    ] + extra_imports)
    extra_imports.clear()
    return i


def go_options(extra_options=[]):
    """Returns the default options and extra ones if specified."""
    if isinstance(extra_options, str):
        extra_options = [extra_options]
    o = ','.join([
        'Mgoogle/api/annotations.proto=google.golang.org/genproto/googleapis/api/annotations',
        'Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types',
        'Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types',
        'Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types'
    ] + extra_options)
    extra_options.clear()
    return o


def glob(pattern: str, dir=cwd()):
    """Returns file names that match the given pattern. If the second parameter is not
    specified, files are looked up in cwd."""
    return _glob.glob(os.path.join(dir, pattern))


def arch(when_darwin=None):
    if sys.platform.startswith('darwin') and when_darwin is not None:
        return when_darwin
    return machine()


def platform(when_darwin='darwin'):
    """Returns the platform name as string.
    Arguments:
        when_darwin: the preferred return value when sys.platform is darwin"""
    if sys.platform.startswith('linux'):
        return 'linux'
    elif sys.platform.startswith('darwin'):
        return when_darwin
    else:
        return sys.platform


def join_paths(parent: str, elements: list):
    """Joins each element path in elements with parent as its parent.
    Arguments:
        parent: A relative or absolute path
        elements: A list of path elements. Each element can either be a list of strings
        or a string. If string, it is treated as a list of strings of length 1."""
    all = []
    for paths in elements:
        if isinstance(paths, str):
            paths = [paths]
        p = parent
        for path in paths:
            p = os.path.join(p, path)
        all.append(p)
    return all


def _parent_dir(path: str):
    """Utility function to get the absolute path of the parent of the given path"""
    return os.path.abspath(os.path.join(path, os.pardir))


def _call_process(args: list, env=None, cwd=None, raise_err: bool = True):
    """_call_process directy calls subprocess.call
    Arguments:
       args: program name and command line arguments
       env: env of the subprocess. If left blank, it is the default env.
       cwd: cwd of the suprocess. If left blank, it is the default cwd.
       raise_err: If true and the process ends with a non-zero code, an exception is raised. Default is True."""
    code = subprocess.call(
        args, env=env, cwd=cwd, stdout=sys.stdout)
    if code != 0 and raise_err:
        raise Exception(f'"{" ".join(args)}" failed with code {code}')
    return code


def _read_config():
    """Opens the config file at the root of the project and loads
    it to _config variable. If file is not found, program exits."""
    try:
        config_path = os.path.join(root_dir(), 'protos.json')
        with open(config_path) as file:
            return json.load(file)
    except OSError:
        # if current repo is apps, root/apis/protos.json will work
        try:
            config_path_apps = os.path.join(root_dir(), 'apis', 'protos.json')
            with open(config_path_apps) as file:
                return json.load(file)
        except OSError:
            print(
                f'config file must be at: {config_path} or {config_path_apps} if working in apps repo')
            exit(1)


def _check_and_download_grpc_java(version: str, clear_cache: bool):
    """Checks if grpc-java plugin is downloaded. If it does not exist, or clear_cache is true,
    it is downloaded. In either case, the import path for the plugin is returned."""
    a = arch()
    p = platform('osx')
    download_to = os.path.join(
        root_dir(), 'bin', 'protoc-gen-grpc-java', 'protoc-gen-grpc-java')
    download_dir = os.path.join(root_dir(), 'bin', 'protoc-gen-grpc-java')
    if clear_cache or not os.path.exists(download_to):
        _call_process(
            ['mkdir', '-p', os.path.abspath(os.path.join(download_to, os.pardir))])
        print(
            f'downloading protoc-gen-grpc-java plugin ({version}-{p}-{a}) to {download_to}')
        url = f'https://repo1.maven.org/maven2/io/grpc/protoc-gen-grpc-java/{version}/protoc-gen-grpc-java-{version}-{p}-{a}.exe'
        _call_process(
            ['curl', '-L', '--url', url, '--output', download_to, '--silent'])
        _call_process(['chmod', '+x', download_to])

    return download_dir


def _check_and_build_swift(version: str, clear_cache: bool):
    root = root_dir()
    plugin_path = os.path.join(
        root, 'bin', 'protoc-gen-swift', 'protoc-gen-swift')
    plugin_dir = os.path.join(root, 'bin', 'protoc-gen-swift')
    clone_dir = os.path.join(root, 'bin', 'swift-protobuf')

    if clear_cache or not os.path.exists(plugin_path):
        _call_process(['mkdir', '-p', plugin_dir])
        # clone repository
        if os.path.exists(clone_dir):
            shutil.rmtree(clone_dir)
        _call_process(
            ['git', 'clone', 'https://github.com/apple/swift-protobuf.git', clone_dir])
        print(f'checkout swift-protobuf to tags/{version}')
        _call_process(['git', 'checkout', f'tags/{version}'], cwd=clone_dir)

        # build
        swift_dir = subprocess.check_output(
            ['swift', 'build', '-c', 'release', '--show-bin-path'], cwd=clone_dir, encoding=sys.getdefaultencoding()).rstrip()
        print(
            f'building protoc-gen-swift with version {version} to {swift_dir}')
        _call_process(['swift', 'build', '-c', 'release'], cwd=clone_dir)

        # copy to plugin_path
        shutil.copy(os.path.join(swift_dir, 'protoc-gen-swift'), plugin_path)
        shutil.rmtree(clone_dir)

    return plugin_dir


def _check_and_build_swift_grpc(version: str, clear_cache: bool):
    root = root_dir()
    plugin_path = os.path.join(
        root, 'bin', 'protoc-gen-grpc-swift', 'protoc-gen-grpc-swift')
    plugin_dir = os.path.join(root, 'bin', 'protoc-gen-grpc-swift')
    clone_dir = os.path.join(root, 'bin', 'grpc-swift')

    if clear_cache or not os.path.exists(plugin_path):
        _call_process(['mkdir', '-p', plugin_dir])
        # clone repository
        if os.path.exists(clone_dir):
            shutil.rmtree(clone_dir)
        _call_process(
            ['git', 'clone', 'https://github.com/grpc/grpc-swift.git', clone_dir])
        print(f'checkout swift-protobuf to tags/{version}')
        _call_process(['git', 'checkout', f'tags/{version}'], cwd=clone_dir)

        # build
        swift_dir = subprocess.check_output(
            ['swift', 'build', '-c', 'release', '--show-bin-path'], cwd=clone_dir, encoding=sys.getdefaultencoding()).rstrip()
        print(
            f'building protoc-gen-grpc-swift with version {version} to {swift_dir}')
        _call_process(['swift', 'build', '-c', 'release',
                      '--product', 'protoc-gen-grpc-swift', ], cwd=clone_dir)

        # copy to plugin_path
        shutil.copy(os.path.join(
            swift_dir, 'protoc-gen-grpc-swift'), plugin_path)
        shutil.rmtree(clone_dir)

    return plugin_dir


def _check_and_build_or_download_go_plugin(plugin: str, config: Any, clear_cache: bool):
    full_name = f'protoc-gen-{plugin}'
    root = root_dir()
    plugin_path = os.path.join(root, 'bin', full_name, full_name)
    plugin_dir = os.path.join(root, 'bin', full_name)

    if clear_cache or not os.path.exists(plugin_path):
        _call_process(['mkdir', '-p', plugin_dir])
        if plugin == 'grpc-gateway':
            _download_gateway(os.path.join(
                plugin_dir, full_name), config['versions'][plugin])
        else:
            _call_process(['obm', 'build', full_name], cwd=root_dir())

    return plugin_dir


def _prepare_dependencies(plugins: list):
    """Builds and if necessary downloads the given plugins and protoc.
    Returns the paths to the executables of the plugins."""
    clear_cache = os.getenv('PROTO_CLEAR_CACHE', False)
    if clear_cache == 'true':
        clear_cache = True
    else:
        clear_cache = False

    config = _read_config()
    paths = []
    # create the bin folder if not exists
    bin_folder = os.path.join(root_dir(), 'bin')
    if not os.path.exists(bin_folder):
        _call_process(['mkdir', '-p', bin_folder])

    # check protoc
    protoc_version = config['versions']['protoc']
    protoc_path = os.path.join(bin_folder, 'protoc-bin')

    if clear_cache or (not os.path.exists(os.path.join(protoc_path, 'bin', 'protoc'))):
        # download protoc
        if not os.path.exists(protoc_path):
            _call_process(['mkdir', '-p', protoc_path])

        _download_protoc(protoc_path, protoc_version)
        paths.append(os.path.join(protoc_path, 'bin'))
    else:
        # already downloaded
        paths.append(os.path.join(protoc_path, 'bin'))

    # check plugins
    plugin_path = ''
    # java_out is implemented within protoc, no external plugin is needed
    prepared = ['', 'java']
    for plugin in plugins:
        if plugin in prepared:
            continue
        if platform() != 'darwin' and (plugin == 'swift' or plugin == 'grpc-swift'):
            continue
        if plugin == 'grpcio-tools':
            _check_grpcio_tools(clear_cache)
            # save the path to the python3 binary
            plugin_path = _parent_dir(sys.executable)
        elif plugin == 'grpc-java':
            plugin_path = _check_and_download_grpc_java(
                config["versions"]["grpc-java"], clear_cache)
        elif plugin == 'swift':
            plugin_path = _check_and_build_swift(
                config['versions']['swift'], clear_cache)
        elif plugin == 'grpc-swift':
            plugin_path = _check_and_build_swift_grpc(
                config['versions']['grpc-swift'], clear_cache)
        else:
            # go plugins
            plugin_path = _check_and_build_or_download_go_plugin(
                plugin, config, clear_cache)

        paths.append(plugin_path)
        prepared.append(plugin)
    return paths


def _download_gateway(save_to: str, version: str):
    """Downloads the the specified version of grpc-gateway's compiled binary under the given directory."""
    p = platform()
    a = arch("x86_64")
    url = f'https://github.com/grpc-ecosystem/grpc-gateway/releases/download/{version}/protoc-gen-grpc-gateway-{version}-{p}-{a}'
    print(f'downloading grpc-gateway ({version}-{p}-{a})')
    _call_process(['curl', '-L', '--url', url,
                   '--output', save_to, '--silent'])
    _call_process(['chmod', '774', save_to])


def _download_protoc(protoc_path: str, version: str):
    p = platform('osx')
    a = arch('universal_binary')
    """Downloads the specified version of protoc under the given directory and extracts the zip archive."""
    url = f'https://github.com/protocolbuffers/protobuf/releases/download/{version}/protoc-{version.replace("v", "")}-{p}-{a}.zip'
    zip_path = os.path.join(protoc_path, 'protoc.zip')
    print(f"downloading protoc ({version}-{p}-{a})")
    _call_process(
        ['curl', '-L', '--url', url, '--output', zip_path, '--silent'])
    _call_process(
        ['unzip', '-q', '-o', zip_path, '-d', protoc_path])
    _call_process(['rm', '-f', zip_path])


def _check_grpcio_tools(clear_cache: bool):
    out = subprocess.check_output(['pip3', 'list'], encoding='utf-8')
    if 'grpcio-tools' not in out or clear_cache:
        print(f'downloading grpcio-tools')
        _call_process(['pip3', 'install', 'grpcio-tools'])


class ProtobufData:
    """It is for saving the data of protobuf functions for using at the exit"""

    def __init__(self, plugin: str, files: list, args: list, protoc_args: list, proto_path: str, out: str, python_out: bool, cwd: str, gofmt: bool):
        self.plugin = plugin
        self.files = files.copy()
        self.plugin_args = args.copy()
        self.protoc_args = protoc_args.copy()
        self.proto_path = proto_path
        self.out = out
        self.python_out = python_out
        self.cwd = cwd
        self.gofmt = gofmt


def _run_protoc():
    """_run_protoc finds the required dependencies for registered plugins and calls
    protoc for each of them with its arguments."""
    check_cache = os.getenv('PROTO_CLEAR_CACHE')
    if check_cache == 'true':
        check_cache = True
    else:
        check_cache = False

    # find required dependencies
    deps = []
    for cmd in _registered:
        deps.append(cmd.plugin)
    env = {
        'PATH': os.pathsep.join(_prepare_dependencies(deps))
    }

    for cmd in _registered:
        current_dir = cmd.cwd
        if not current_dir:
            current_dir = cwd()
        outdir = os.path.join(current_dir, cmd.out)
        if not os.path.exists(outdir):
            _call_process(['mkdir', '-p', outdir])

        if cmd.python_out:
            _call_process([
                'python3', '-m', 'grpc_tools.protoc',
                f'--proto_path={cmd.proto_path}',
                f'--python_out={cmd.out}',
                f'--grpc_python_out={cmd.out}'
            ] + cmd.files, env=env, cwd=current_dir)
        else:
            if cmd.proto_path != '':
                cmd.protoc_args.append(f'--proto_path={cmd.proto_path}')
            if cmd.plugin != '':
                if platform() != 'darwin' and (cmd.plugin == 'swift' or cmd.plugin == 'grpc-swift'):
                    continue
                if len(cmd.plugin_args) == 0:
                    cmd.protoc_args.append(f'--{cmd.plugin}_out={cmd.out}')
                else:
                    cmd.protoc_args.append(
                        f'--{cmd.plugin}_out={",".join(cmd.plugin_args)}:{cmd.out}')
            _call_process(
                ['protoc'] + cmd.protoc_args + cmd.files, env=env, cwd=current_dir)
            if cmd.plugin == 'otsimo-auth' and cmd.gofmt:
                auth_files = []
                for root, dirs, files in os.walk(os.path.join(current_dir, cmd.out)):
                    for f in files:
                        if f.endswith('.auth.go'):
                            auth_files.append(os.path.join(root, f))

                _call_process(['go', 'fmt'] +
                              auth_files, raise_err=True)


def protobuf(files: list, plugin: str = '', args: list = [], protoc_args: list = [],
             proto_path: str = '', out: str = os.curdir, python_out=False, options_go=None, cwd: str = '', gofmt: bool = False):
    """At each call, registers a call to protoc, adding default arguments if necessary.
    Arguments:
        plugin: 
            Name of the plugin without the preceding 'protoc-gen-'. If left empty,
            no plugin is checked and added to protoc.
        files: 
            List of file paths.
        args: 
            List of plugin arguments. If only one arg is to bespecified, it can be given as str.
        protoc_args: 
            List of protoc arguments. This is the way to directly pass parameters to protoc.
        proto_path:
            The import path for protoc. Default is the default imports, given by import_path().
        out:
            Output directory. Can be specified as relative path. Default is cwd.
        python_out:
            If true, it will output python protobuf. Default is false.
        options_go:
            The options to be passed to plugin. DEfault is the default options, given by go_options()."""

    if len(_registered) == 0:
        # if this is the first call to protobuf,
        # _run_protoc should be registered now
        atexit.register(_run_protoc)

    if proto_path == '':
        proto_path = import_path()

    if isinstance(args, str):
        args = [args]

    if python_out:
        plugin = 'grpcio-tools'
    elif plugin in ('java', 'swift', 'grpc-swift'):
        pass
    else:
        if options_go != None:
            if isinstance(options_go, list):
                args += options_go
            elif isinstance(options_go, str):
                args.append(options_go)
        else:
            # use default options
            args.append(go_options())

    _registered.append(ProtobufData(
        plugin, files, args, protoc_args, proto_path, out, python_out, cwd, gofmt
    ))

    # clean-up
    args.clear()
    protoc_args.clear()
    cwd = ''


if __name__ == '__main__':
    # add the path of this file to env so that called script can find it
    protos_path = cwd()
    env = os.environ.copy()
    if env.get('PYTHONPATH') == None:
        env['PYTHONPATH'] = protos_path
    else:
        env['PYTHONPATH'] += f'{os.path.pathsep}{protos_path}'

    try:
        i = sys.argv.index('--clear-cache')
        del sys.argv[i]
        env['PROTO_CLEAR_CACHE'] = 'true'
    except ValueError:
        env['PROTO_CLEAR_CACHE'] = 'false'

    for script in sys.argv:
        if script == sys.argv[0]:
            continue
        try:
            _call_process(['python3', script], env=env)
        except KeyboardInterrupt:
            print('received keyboard interrupt, finishing program')
            exit(1)
