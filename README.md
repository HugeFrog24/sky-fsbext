# Sky: CotL Fmod Extractor

This script extracts audio data from sound banks in the assets folder of the video game Sky: Children of the Light and saves them as .wav files using the vgmstream audio decoder.

The extracted audio files can be used to listen to the game's audio outside of the game environment, for example, with a regular audio player or for other non-commercial purposes.

## Prerequisites
- vgmstream-cli
- One of the following:
  - An unpacked Sky APK with the sound banks you wish to extract (usually located at `/path/to/apk/assets/Data/Audio/Fmod/fmodandroid/`)
  - Sky: Children of the Light installed via Steam (Windows only - auto-detection supported)
- 7 GB minimum free disk space

## Usage

The application supports automatic Steam detection on Windows. If you have Sky: Children of the Light installed via Steam, the program will automatically detect and use the game's audio files when no bank files are found in the input directory.

### For Developers
1. Ensure that Go 1.23.2 or higher is installed on your system.
2. Clone the repository and navigate to the project directory.
3. Run the program using `go run .` with optional command-line arguments:
    - `-i` or `--input-dir` to specify the path to the input directory (default is `in`).
    - `-o` or `--output-dir` to specify the path to the output directory (default is `out`).
    - `-p` or `--vgmstream-path` to provide the path to the `vgmstream-cli` executable (default is `vgmstream-win64/vgmstream-cli.exe`).
    - `-c` or `--compression-ratio` to specify the compression ratio used for calculating disk space requirements (default is 8.0).
    - `-v` or `--verbose` to enable verbose output.
    - `-w` or `--workers` to set the number of concurrent workers (default is 4).
    - `--version` to print the program version.
4. Alternatively, build the program using `go build -o build/sky-fsbext` and run the resulting executable from the `build` directory.

### For End Users
1. Download the latest release binary from the [Releases](https://github.com/HugeFrog24/sky-fsbext/releases) page.
2. Ensure that `vgmstream-cli` is installed and accessible from the command line.
3. Run the program with optional command-line arguments:
    - `-i` or `--input-dir` to specify the path to the input directory (default is `in`).
    - `-o` or `--output-dir` to specify the path to the output directory (default is `out`).
    - `-p` or `--vgmstream-path` to provide the path to the `vgmstream-cli` executable (default is `vgmstream-win64/vgmstream-cli.exe`).
    - `-c` or `--compression-ratio` to specify the compression ratio used for calculating disk space requirements (default is 8.0).
    - `-v` or `--verbose` to enable verbose output.
    - `-w` or `--workers` to set the number of concurrent workers (default is 4).
    - `--version` to print the program version.
4. Wait for the program to finish processing.
5. The extracted audio files will be located in the output directory.

### Steam Auto-Detection (Windows Only)
- If no `.bank` files are found in the input directory, the program will automatically attempt to detect Sky: Children of the Light installed via Steam.
- The program searches the Windows registry for the Steam installation path and locates audio files in the game's asset directories.
- This feature works seamlessly without requiring manual file copying or path configuration.

## Configuration
- The program logs its progress to `fsbext.log`.
- The directory structure for the extracted audio files is as follows:
    - Music
    - SFX
    - Other

## Screenshots

<table>
  <tr>
    <td valign="top" style="text-align: center;">
      <img src="assets/screenshots/scr-cap-terminal.png" alt="Demo" /><br />
      Demo
    </td>
    <td valign="top" style="text-align: center;">
      <img src="assets/screenshots/scr-cap-audio-wave-spectrum.png" alt="Extracted Audio Waveform and Spectrum" /><br />
      Extracted Audio Waveform and Spectrum
    </td>
  </tr>
</table>

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.
