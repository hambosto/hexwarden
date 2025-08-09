# Test Data

This directory contains test data files used by the HexWarden test suite.

## Files

- `sample.txt` - Small text file for basic testing
- `binary.dat` - Binary data file for testing binary operations
- `large.txt` - Larger text file for performance testing
- `empty.txt` - Empty file for edge case testing

## Usage

These files are used by various test functions to provide consistent test data across different test scenarios. The files are designed to test different aspects of the encryption/decryption system:

- Text files test string handling
- Binary files test raw byte operations
- Large files test performance and memory usage
- Empty files test edge cases

## Maintenance

When adding new test data files:
1. Document them in this README
2. Keep file sizes reasonable for CI/CD environments
3. Use descriptive names that indicate the test purpose
4. Consider both positive and negative test cases