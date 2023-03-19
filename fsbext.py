import argparse
import platform
import shutil
import subprocess
import sys
import logging
from pathlib import Path

__author__ = "Tibik"
__version__ = "1.0.1"

# Define the minimum required disk space (in bytes)
MIN_DISK_SPACE = 7 * 1024 * 1024 * 1024


class ConsoleFilter(logging.Filter):
    def filter(self, record):
        return not getattr(record, "log_to_file_only", False)


# Set up root logger
root_logger = logging.getLogger()
root_logger.setLevel(logging.DEBUG)

# Add file handler to root logger
file_handler = logging.FileHandler("fsbext.log", mode="w")
file_handler.setFormatter(logging.Formatter('%(asctime)s %(levelname)s: %(message)s', datefmt='%Y-%m-%d %H:%M:%S'))
root_logger.addHandler(file_handler)

# Add console handler to root logger
console_handler = logging.StreamHandler(sys.stdout)
console_handler.setLevel(logging.INFO)
console_formatter = logging.Formatter('%(message)s')
console_handler.setFormatter(console_formatter)
console_handler.addFilter(ConsoleFilter())  # Add the custom filter to the console handler
root_logger.addHandler(console_handler)

LOGGER_PADDING = '=' * 10


def check_disk_space():
    # Check available disk space
    free_space = shutil.disk_usage(".").free
    if free_space < MIN_DISK_SPACE:
        root_logger.warning(
            f"Less than {MIN_DISK_SPACE / (1024 * 1024 * 1024):.2f} GB of disk space available "
            f"({free_space / (1024 * 1024 * 1024):.2f} GB)"
        )


def create_directory_structure(output_dir: Path):
    (output_dir / "Music").mkdir(parents=True, exist_ok=True)
    (output_dir / "SFX").mkdir(parents=True, exist_ok=True)
    (output_dir / "Other").mkdir(parents=True, exist_ok=True)
    root_logger.info("Created directory structure.")


def remove_empty_directories(output_dir: Path):
    for dir_path in output_dir.glob('**/*'):
        if dir_path.is_dir() and not any(dir_path.iterdir()):
            shutil.rmtree(dir_path)
            root_logger.info(f"Removed empty directory: {dir_path}", extra={"log_to_file_only": True})


def extract_and_move_files(args, bank_files):
    extracted_files = 0
    for bank_file in bank_files:
        # Determine the output directory
        if bank_file.startswith("Music_"):
            bank_dir = args.output_dir / "Music" / bank_file[:-5]
        elif bank_file.startswith("SFX_"):
            bank_dir = args.output_dir / "SFX"
        else:
            bank_dir = args.output_dir / "Other" / bank_file[:-5]

        # Create the output directory if it doesn't exist
        bank_dir.mkdir(parents=True, exist_ok=True)
        root_logger.info(f"Created output directory: {bank_dir}", extra={"log_to_file_only": True})

        # Extract the bank file to WAV files
        status = None
        err_reason = ""
        try:
            subprocess.run(
                [
                    args.vgmstream_path, args.input_dir / bank_file,
                    "-o", bank_dir / "?n.wav", "-S", "0"
                ], check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
            )
        except subprocess.CalledProcessError as e:
            status = False
            err_reason = e
        else:
            status = True
            extracted_files += 1
        finally:
            root_logger.info(
                f"Processing {bank_file}: {'OK!' if status else 'ERR!'}"
            )
            if not status:
                root_logger.error(f"An error occurred while extracting {bank_file}: {err_reason}",
                                  extra={"log_to_file_only": True})

            # Check if the output directory is empty and remove it if it is
            if not any(bank_dir.iterdir()):
                shutil.rmtree(bank_dir)
                root_logger.info(f"Removed empty directory: {bank_dir}", extra={"log_to_file_only": True})

    return extracted_files


def main(args):
    version_str = f"SKY-FSBEXT version: {__version__} by {__author__}"
    root_logger.info(f"{LOGGER_PADDING} {version_str} {LOGGER_PADDING}")
    root_logger.info(f"Operating system: {platform.platform()}")
    if args.version:
        print(version_str)
        exit()
    check_disk_space()

    # Log input and output directories
    root_logger.info(f"Input directory: {args.input_dir}")
    root_logger.info(f"Output directory: {args.output_dir}")

    create_directory_structure(args.output_dir)

    # Check if the "in" directory exists and rebuild it if necessary
    if not args.input_dir.is_dir():
        args.input_dir.mkdir(parents=True)
        root_logger.warning("Input directory not found - rebuilding")

    # Search for .bank files in the in directory
    bank_files = [f for f in args.input_dir.glob("*.bank")]

    if not bank_files:
        root_logger.warning("No sound banks found in input directory")
    else:
        root_logger.info(f"Found {len(bank_files)} sound bank(s) in input directory")

    # Check if the vgmstream executable is present and get its version number
    if not args.vgmstream_path.is_file():
        root_logger.error("vgmstream-cli executable not found")
        exit(1)

    extracted_files = extract_and_move_files(args, bank_files)

    if extracted_files > 0:
        root_logger.info(f"Successfully extracted {extracted_files} bank file(s)")
    else:
        root_logger.warning("No sound banks were extracted")

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
    parsed_args = parser.parse_args()
    main(args=parsed_args)
    root_logger.info(f"{LOGGER_PADDING} Done, program exiting. {LOGGER_PADDING}")
else:
    print("This script is intended to be run from the command line. "
          "Please run 'python fsbext.py --help' for usage information.")
