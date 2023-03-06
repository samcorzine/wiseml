import os
import tarfile
import tempfile
def create_tar_file(source_files, target=None):
    if target:
        filename = target
    else:
        _, filename = tempfile.mkstemp()

    with tarfile.open(filename, mode="w:gz", dereference=True) as t:
        for sf in source_files:
            # Add all files from the directory into the root of the directory structure of the tar
            t.add(sf, arcname=os.path.basename(sf))
    return filename

def tar_from_dir(dir_path: str, target=None) -> str:
    """Walk the directory to get all the filepaths and then pass them to create_tar_file"""

    source_files = []
    for root, _, files in os.walk(dir_path):
        for f in files:
            source_files.append(os.path.join(root, f))
    return create_tar_file(source_files, target)