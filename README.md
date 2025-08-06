<div align="center">
  <h1>Hexwarden</h1>
  <p>
    <a href="https://goreportcard.com/report/github.com/hambosto/hexwarden"><img src="https://goreportcard.com/badge/github.com/hambosto/hexwarden" alt="Go Report Card"></a>
    <a href="https://github.com/hambosto/hexwarden/releases"><img src="https://img.shields.io/github/v/release/hambosto/hexwarden" alt="Latest Release"></a>
    <a href="https://github.com/hambosto/hexwarden/blob/main/LICENSE"><img src="https://img.shields.io/github/license/hambosto/hexwarden" alt="License"></a>
  </p>
</div>

Hexwarden is a command-line tool that makes file encryption simple and secure. It uses strong encryption to protect your files, and it's designed to be easy to use, even if you're not an expert.

Whether you need to protect sensitive documents, secure backups, or just keep your personal files private, Hexwarden has you covered.

## Features

- **Strong Encryption**: Your files are protected with AES-256-GCM, a modern and secure encryption standard.
- **User-Friendly**: An interactive command-line interface guides you through the process of encrypting and decrypting files.
- **Cross-Platform**: Works on Windows, macOS, and Linux.
- **Secure Deletion**: Option to securely overwrite and delete original files after encryption/decryption.
- **Open Source**: The code is open for anyone to inspect and contribute to.

## Getting Started

### Prerequisites

- **Go**: Version 1.19 or higher (only if building from source).

### Installation

You can either download a pre-built binary or build it from the source.

**From Releases (Recommended)**

1.  Go to the [Releases](https://github.com/hambosto/hexwarden/releases) page.
2.  Download the latest binary for your operating system.
3.  (On macOS/Linux) Make the file executable:
    ```bash
    chmod +x hexwarden
    ```

**From Source**

1.  Clone the repository:
    ```bash
    git clone https://github.com/hambosto/hexwarden.git
    ```
2.  Navigate to the project directory:
    ```bash
    cd hexwarden
    ```
3.  Build the project:
    ```bash
    go build -o hexwarden cmd/hexwarden/main.go
    ```

## Usage

1.  Run Hexwarden from your terminal:
    ```bash
    ./hexwarden
    ```
2.  Choose whether you want to **Encrypt** or **Decrypt** a file.
3.  Select the file you want to process from the list.
4.  Enter a strong password to secure your file.

That's it! Hexwarden will take care of the rest. Encrypted files will be saved with a `.hex` extension, and decrypted files will be restored to their original names.

## How It Works

Hexwarden is built with security and simplicity in mind. Here's a quick look at what happens under the hood:

1.  **Password-Based Key Derivation**: When you enter a password, Hexwarden uses a Key Derivation Function (KDF) to turn it into a strong encryption key. This makes it much harder for someone to guess your password.
2.  **Compression**: Before encrypting, your file is compressed. This can make the final encrypted file smaller.
3.  **Encryption**: The compressed data is then encrypted using AES-256-GCM. This not only secures your data but also ensures that it hasn't been tampered with.
4.  **Secure Header**: Important information, like the original filename and the salt used for key derivation, is stored in a secure header.
5.  **Chunk Processing**: For larger files, Hexwarden processes them in smaller chunks, which is more memory-efficient.

## Security

Hexwarden is designed with a focus on security. Here are some of the key security features:

- **AES-256-GCM**: A strong, authenticated encryption algorithm that provides both confidentiality and integrity.
- **Argon2id**: A modern, secure Key Derivation Function used to protect your password.
- **Cryptographically Secure Randomness**: Used for generating nonces and salts to ensure that every encryption is unique.

While Hexwarden is designed to be secure, it's important to follow good security practices:
- Use a strong, unique password.
- Keep your password safe and don't share it with anyone.

## Contributing

We welcome contributions from the community! If you'd like to help improve Hexwarden, please feel free to:

- Report bugs or suggest new features by opening an [issue](https://github.com/hambosto/hexwarden/issues).
- Submit a pull request with your improvements.

Please read our [Contributing Guidelines](CONTRIBUTING.md) for more information.

## License

Hexwarden is open-source software licensed under the [MIT License](LICENSE).