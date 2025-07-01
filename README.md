# MHTML File Extractor

A lightweight GUI application built with [Gio](https://gioui.org) to parse and extract resources from MHTML (.mhtml, .mht) files. It allows users to view raw HTML content, toggle external script fetching, and extract embedded resources (e.g., images, CSS, JavaScript) to a specified directory.

## Features

- **Parse MHTML Files**: Load and parse MHTML files to extract embedded resources and HTML content.
- **Raw Source View**: Display the raw HTML content in a read-only editor.
- **Configurable External Fetching**: Toggle fetching of external JavaScript files via a checkbox, with concurrent downloads using a worker pool.
- **Resource Extraction**: Select and extract resources (e.g., images, scripts) to a user-specified output directory.
- **Dark/Light Mode**: Switch between dark and light themes for better usability.
- **Cross-Platform**: Supports Windows and Linux, with macOS support for users with Xcode installed.

## Prerequisites

- **Go**: Version 1.21 or later ([download](https://golang.org/dl)).
- **System Dependencies**:
  - **Linux**: `libgtk-3-dev`, `libgles2-mesa-dev`, `libx11-dev`, `xorg-dev`, `gcc` (see [Gio Linux setup](https://gioui.org/doc/install/linux)).
  - **Windows**: No additional dependencies required (Edge WebView2 runtime is included in Windows 10/11).
  - **macOS**: Xcode (for CGO and Gio’s Objective-C bindings, see [Gio macOS setup](https://gioui.org/doc/install/macos)).
- **Optional**: [UPX](https://upx.github.io) for binary compression to reduce executable size.

## Installation

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/yourusername/mhtml-extractor.git
   cd mhtml-extractor
   ```

2. **Initialize Go Module**:
   ```bash
   go mod init github.com/yourusername/mhtml-extractor
   ```

3. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

## Building the Executable

### Linux

1. **Install System Dependencies**:
   ```bash
   sudo apt-get update
   sudo apt-get install -y gcc pkg-config libwayland-dev libx11-dev libx11-xcb-dev libxkbcommon-x11-dev libgles2-mesa-dev libegl1-mesa-dev libffi-dev libxcursor-dev libvulkan-dev
   ```

2. **Build the Executable**:
   ```bash
   go build -ldflags="-s -w" -o mhtml-extractor
   ```

3. **Compress with UPX** (optional, for smaller binary):
   ```bash
   sudo apt-get install -y upx-ucl
   upx --best mhtml-extractor
   ```

4. **Run**:
   ```bash
   ./mhtml-extractor
   ```

A GitHub Actions workflow (`.github/workflows/build-linux.yml`) automates building for Linux (`amd64`). Check the [Actions tab](https://github.com/yourusername/mhtml-extractor/actions) for pre-built binaries (`mhtml-extractor-linux`).

### Windows

1. **Build the Executable**:
   ```bash
   go build -ldflags="-s -w" -o mhtml-extractor.exe
   ```

2. **Compress with UPX** (optional):
   - Download UPX from [upx.github.io](https://upx.github.io) and add it to your PATH.
   ```bash
   upx --best mhtml-extractor.exe
   ```

3. **Run**:
   ```bash
   .\mhtml-extractor.exe
   ```

### macOS

**Note**: Building on macOS requires Xcode due to Gio’s Objective-C bindings. If you don’t have Xcode, see the [Sidenote for macOS Users](#sidenote-for-macos-users) below.

1. **Install Xcode**:
   - Download Xcode from the [Mac App Store](https://apps.apple.com/us/app/xcode/id497799835) or [Apple Developer](https://developer.apple.com/xcode/).
   - Install Command Line Tools:
     ```bash
     xcode-select --install
     ```

2. **Build the Executable**:
   ```bash
   go build -ldflags="-s -w" -o mhtml-extractor
   ```

3. **Compress with UPX** (optional):
   ```bash
   brew install upx
   upx --best mhtml-extractor
   ```

4. **Run**:
   ```bash
   ./mhtml-extractor
   ```

## Usage

1. Launch the application (`mhtml-extractor` or `mhtml-extractor.exe`).
2. Click **Browse** to select an MHTML (.mhtml, .mht) file.
3. Toggle **Fetch External Scripts** to include external JavaScript (downloaded concurrently).
4. View raw HTML in the **Raw Source** section.
5. Select resources in the **Embedded Resources** table and click **Extract Selected** to save them to the output directory (defaults to a folder named after the MHTML file).
6. Click **Change Output Dir** to set a custom output directory.
7. Toggle **Mode** (🌓) to switch between dark and light themes.

## Binary Size Optimization

The executable is optimized for size using:
- `-ldflags="-s -w"`: Strips debug symbols and DWARF tables (~20–33% reduction).
- UPX compression: Reduces binary to ~2–4 MB (adds ~170–180 ms startup delay).

Unoptimized size: ~10–15 MB. To further reduce size, consider replacing `zenity` (requires CGO) or `goquery` with lighter alternatives.

## Sidenote for macOS Users

Due to Gio’s dependency on Xcode, building for macOS is challenging without a macOS environment. If you’ve successfully built the `mhtml-extractor` binary for macOS (`amd64`), please share it with the project maintainer via a fast file transfer service (e.g., [WeTransfer](https://wetransfer.com), [TransferNow](https://www.transfernow.net)). Send the link to [your-email@example.com](mailto:your-email@example.com) or open an issue/pull request. Your binary will be uploaded to the [GitHub Releases](https://github.com/yourusername/mhtml-extractor/releases) to benefit other macOS users.

## Contributing

Contributions are welcome! Please:
- Open issues for bugs or feature requests.
- Submit pull requests with clear descriptions.
- For macOS builds, share binaries as noted above.

## License

[MIT License](LICENSE)
