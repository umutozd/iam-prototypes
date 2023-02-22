from protos import protobuf, glob, cwd
import os

files = glob('*.proto')
out = '.'
gwout = './gw'

protobuf(
    plugin='gofast',
    args='plugins=grpc',
    out=out,
    files=files
)
