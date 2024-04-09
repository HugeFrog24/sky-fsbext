import argparse
import platform
import shutil
import subprocess
import sys
import logging
from pathlib import Path

__author__ = "Tibik"
__version__ = "1.0.6"


class ConsoleFilter(logging.Filter):
    def filter(self, record):
        return not getattr(record, "log_to_file_only", False)


# Set up module-specific logger
logger = logging.getLogger(__name__)

def setup_logging(verbose: bool, apply_console_filter: bool = True):
    """
    Set up logging for the application.
    
    :param verbose: If True, set log level to DEBUG, else INFO.
    :param apply_console_filter: Apply ConsoleFilter if True.
    """
    log_level = logging.DEBUG if verbose else logging.INFO
    logger.setLevel(log_level)

    # Add file handler
    file_handler = logging.FileHandler("fsbext.log", mode="w")
    file_formatter = logging.Formatter('%(asctime)s %(levelname)s: %(message)s', datefmt='%Y-%m-%d %H:%M:%S')
    file_handler.setFormatter(file_formatter)
    logger.addHandler(file_handler)

    # Add console handler
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.setLevel(log_level)
    console_formatter = logging.Formatter('%(message)s')
    console_handler.setFormatter(console_formatter)

    if apply_console_filter:
        console_handler.addFilter(ConsoleFilter())  # Add the custom filter to the console handler

    logger.addHandler(console_handler)

LOGGER_PADDING = '=' * 10


def get_size_of_dir(directory: Path) -> int:
    """
    Calculate the total size of a directory using pathlib.

    :param directory: Path object representing the directory.
    :return: Total size of the directory in bytes.
    """
    return sum(file.stat().st_size for file in directory.rglob('*') if file.is_file())


def check_disk_space(input_dir: Path, output_dir: Path, compression_ratio: float, disk_usage=shutil.disk_usage, get_size_of_dir=get_size_of_dir):
    """
    Check if there is enough disk space in the output directory based on the size of the input directory and a given compression ratio.

    :param input_dir: Path object representing the input directory.
    :param output_dir: Path object representing the output directory.
    :param compression_ratio: The ratio used to estimate the expected size after extraction.
    :param disk_usage: Function to use for getting disk usage, default is shutil.disk_usage.
    :param get_size_of_dir: Function to use for calculating the size of a directory, default is the get_size_of_dir function defined above.
    """
    # Calculate the size of the input directory in GB
    input_dir_size_gb = get_size_of_dir(input_dir) / (1024 * 1024 * 1024)

    # Calculate the expected space needed for extraction in GB
    expected_size_gb = input_dir_size_gb * compression_ratio

    # Convert expected size to bytes
    expected_size_bytes = expected_size_gb * 1024 * 1024 * 1024

    # Check available disk space in the output directory
    free_space = disk_usage(output_dir).free
    if free_space < expected_size_bytes:
        logger.warning(
            f"Less than {expected_size_gb:.2f} GB of disk space available for extraction "
            f"({free_space / (1024 * 1024 * 1024):.2f} GB free)"
        )


def create_directory_structure(output_dir: Path):
    directories = ["Music", "SFX", "Other"]
    for dir_name in directories:
        try:
            (output_dir / dir_name).mkdir(parents=True, exist_ok=True)
        except PermissionError as e:
            logger.error(f"Failed to create directory {dir_name} due to permission error: {e}")
            continue
        else:
            logger.info(f"Created directory structure for {dir_name}.")


def remove_empty_directories(output_dir: Path):
    for dir_path in output_dir.glob('**/*'):
        if dir_path.is_dir() and not any(dir_path.iterdir()):
            try:
                shutil.rmtree(dir_path)
            except PermissionError:
                logger.error(f"Failed to remove directory: {dir_path}", extra={"log_to_file_only": True})
            else:
                logger.info(f"Removed empty directory: {dir_path}", extra={"log_to_file_only": True})


def extract_and_move_files(args, bank_files: list[Path]):
    extracted_files = 0
    total_files = len(bank_files)
    for i, bank_file in enumerate(bank_files, start=1):
        # Determine the output directory
        if bank_file.name.startswith("Music_"):
            bank_dir = args.output_dir / "Music" / bank_file.stem
        elif bank_file.name.startswith("SFX_"):
            bank_dir = args.output_dir / "SFX" / bank_file.stem
        else:
            bank_dir = args.output_dir / "Other" / bank_file.stem

        # Check if the output directory can be created or exists
        try:
            bank_dir.mkdir(parents=True, exist_ok=True)
            logger.info(f"Created output directory: {bank_dir}", extra={"log_to_file_only": True})
        except PermissionError as e:
            logger.error(f"Failed to create or access directory {bank_dir} due to permission error: {e}")
            continue  # Skip to the next file if the directory cannot be created or accessed

        # Proceed with file extraction
        try:
            subprocess.run(
                [
                    str(args.vgmstream_path), str(bank_file),
                    "-o", str(bank_dir / "?n.wav"), "-S", "0"
                ], check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
            )
            extracted_files += 1
            logger.info(f"Processing file {i} of {total_files}: {bank_file}: OK!")
        except subprocess.CalledProcessError as e:
            logger.error(f"An error occurred while extracting {bank_file}: {e}",
                         extra={"log_to_file_only": True})

        # Check if the output directory is empty and remove it if it is
        if bank_dir.exists() and not any(bank_dir.iterdir()):
            shutil.rmtree(bank_dir)
            logger.info(f"Removed empty directory: {bank_dir}", extra={"log_to_file_only": True})

    return extracted_files


def main(args):
    setup_logging(args.verbose, apply_console_filter=not args.verbose)
    version_str = f"SKY-FSBEXT version: {__version__} by {__author__}"
    logger.info(f"{LOGGER_PADDING} {version_str} {LOGGER_PADDING}")
    logger.info(f"Operating system: {platform.platform()}")
    if args.version:
        logger.info(version_str)
        sys.exit()
    
    check_disk_space(args.input_dir, args.output_dir, args.compression_ratio)

    # Log input and output directories
    logger.info(f"Input directory: {args.input_dir}")
    logger.info(f"Output directory: {args.output_dir}")

    # Check if the "in" directory exists and rebuild it if necessary
    if not args.input_dir.is_dir():
        args.input_dir.mkdir(parents=True)
        logger.warning("Input directory not found - rebuilding")

    # Search for .bank files in the in directory
    bank_files = [f for f in args.input_dir.glob("*.bank")]

    if not bank_files:
        logger.warning("No sound banks found in input directory")
    else:
        logger.info(f"Found {len(bank_files)} sound bank(s) in input directory")
        # Only create directory structure if sound banks are found
        create_directory_structure(args.output_dir)

    # Check if the vgmstream executable is present and get its version number
    if not args.vgmstream_path.resolve().is_file():
        logger.error("vgmstream-cli executable not found")
        sys.exit(1)

    if bank_files:
        extracted_files = extract_and_move_files(args, bank_files)

        if extracted_files > 0:
            logger.info(f"Successfully extracted {extracted_files} bank file(s)")
        else:
            logger.warning("No sound banks were extracted")

        remove_empty_directories(args.output_dir)


if __name__ == "__main__":
    # Define command-line arguments
    parser = argparse.ArgumentParser(
        description="Extracts audio data from sound banks in the assets folder of the video game "
                    "Sky: Children of the Light, and saves them as .wav files using the vgmstream audio decoder. "
                    "The extracted data can be used for game-related purposes, such as listening to the game audio "
                    "outside of the game environment or for other non-commercial purposes."
    )
    parser.add_argument(
        "vgmstream_path", nargs="?", default="vgmstream-win64/vgmstream-cli.exe",
        type=Path, help="Path to vgmstream-cli executable."
    )
    parser.add_argument("-i", "--input-dir", default="in", type=Path, help="Path to the input directory.")
    parser.add_argument("-o", "--output-dir", default="out", type=Path, help="Path to the output directory.")
    parser.add_argument("-v", "--version", action="store_true", help="Prints script version.")
    parser.add_argument("-V", "--verbose", action="store_true", help="Enable verbose output.")
    parser.add_argument(
        "-c", "--compression-ratio", default=8.0, type=float,
        help="Compression ratio used for calculating disk space requirements. Default is 8.0."
    )
    parsed_args = parser.parse_args()
    main(args=parsed_args)
    logger.info(f"{LOGGER_PADDING} Done, program exiting. {LOGGER_PADDING}")
else:
    # Print guidance when this script is imported as a module rather than being run directly.
    print(f"This script is intended to be run from the command line. "
          f"Please run 'python {Path(__file__).name} --help' for usage information.")
