import os
import subprocess
import logging

__author__ = "Tibik"
__version__ = "1.0.0"

# Set up logging
logging.basicConfig(
    filename='fsbext.log', filemode='w', level=logging.DEBUG,
    format='%(asctime)s %(levelname)s: %(message)s', datefmt='%Y-%m-%d %H:%M:%S'
)
# Create the directory structure
os.makedirs("out/Music", exist_ok=True)
os.makedirs("out/SFX", exist_ok=True)
os.makedirs("out/Other", exist_ok=True)
logging.info("Created directory structure")

# Check if the "in" directory exists and rebuild it if necessary
if not os.path.isdir("in"):
    os.makedirs("in")
    logging.warning("Input directory not found - rebuilding")
    print("Input directory not found - rebuilding")

# Search for .bank files in the in directory
bank_files = [f for f in os.listdir("in") if f.endswith(".bank")]
if not bank_files:
    logging.warning("No sound banks found in input directory")
    print("No sound banks found in input directory")
else:
    logging.info(f"Found {len(bank_files)} sound bank(s) in input directory")

    # Check if the vgmstream executable is present and get its version number
    vgmstream_path = os.path.join("vgmstream-win64", "vgmstream-cli.exe")

    # Extract and move the files
    extracted_files = 0
    for bank_file in bank_files:
        # Determine the output directory
        if bank_file.startswith("Music_"):
            out_dir = os.path.join("out", "Music", bank_file[:-5])
        elif bank_file.startswith("SFX_"):
            out_dir = os.path.join("out", "SFX")
        else:
            out_dir = os.path.join("out", "Other", bank_file[:-5])

        # Create the output directory if it doesn't exist
        os.makedirs(out_dir, exist_ok=True)
        logging.info(f"Created output directory: {out_dir}")

        # Extract the bank file to WAV files
        try:
            subprocess.run(
                [
                    vgmstream_path, os.path.join("in", bank_file),
                    "-o", os.path.join(out_dir, "?n.wav"), "-S", "0"
                ], check=True
            )
        except subprocess.CalledProcessError:
            logging.warning(f"Failed to extract {bank_file}")
            print(f"Failed to extract {bank_file}")
        else:
            logging.info(f"Extracted {bank_file} to {out_dir}")
            extracted_files += 1

            # Check if the output directory is empty and remove it if it is
            if not os.listdir(out_dir):
                os.rmdir(out_dir)
                logging.info(f"Removed empty directory: {out_dir}")

    if extracted_files > 0:
        logging.info(f"Successfully extracted {extracted_files} bank file(s)")
        print(f"Successfully extracted {extracted_files} bank file(s)")
    else:
        logging.warning("No sound banks were extracted")
        print("No sound banks were extracted")

    # Remove empty directories
    for root, dirs, files in os.walk("out", topdown=False):
        for directory in dirs:
            dir_path = os.path.join(root, directory)
            if not os.listdir(dir_path):
                os.rmdir(dir_path)
                logging.info(f"Removed empty directory: {dir_path}")
