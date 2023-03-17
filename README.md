# Sky: CotL Fmod Extractor

This script allows you to extract sound banks from the video game Sky: Children of the Light's assets folder.

## Prerequisites
- Python 3
- vgmstream-cli
- 7 GB minimum free disk space

## Usage
1. Extract a Sky APK and locate the sound banks within (usually `/path/to/apk/assets/Data/Audio/Fmod/fmodandroid/`)
2. Paste the sound banks you wish to extract into the script's `in` folder. 
3. Ensure that `vgmstream-cli` is installed and accessible from the command line.
4. Run the script with optional command-line arguments:

  - `-i` or `--input-dir` to specify the path to the input directory (default is `in`)
  - `-o` or `--output-dir` to specify the path to the output directory (default is `out`)
  - `--vgmstream-path` to provide the path to the `vgmstream-cli` executable (default is `vgmstream-win64/vgmstream-cli.exe`)
  - `-v` or `--version` to print the script version.

5. Wait for the script to finish.
6. The extracted audio files will be located in the `out` folder.

## Configuration
- The script logs its progress to `fsbext.log`.
- The directory structure for the extracted audio files is as follows:
  - Music
  - SFX
  - Other

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.