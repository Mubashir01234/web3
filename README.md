# Web3
This project will get block and transaction detail and store information into json file.

## Build
```bash
$ go install
```
The above command will install the project as package in your system with name ```web3```.
## Flags
```bash
web3 --help
Usage of web3:
  -end int
        set ending block. (default 14000000)
  -routines int
        set go routines for processing. (default 5)
  -start int
        set starting block. (default 12000000)
```

## Example
```bash
web3  --start <ex:12000000> --end <ex:14000000> --routines <ex:5>
```