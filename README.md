
# Introduction
The MetaID App Node is an open source backend indexer that accompanies MetaID, which makes it convenient to synchronize MetaID data and asset protocols. MAN is the abbreviation of MetaID App Node. In this document, it specifically refers to the data indexer for UTXO chain applications developed based on the MetaID protocol.

## Main features of MAN
1. Support all UTXO model blockchains
2. Discover MetaID protocol data in block order and transaction order, while supporting data indexing in the memory pool
3. Out of the box, it supports multiple database adapters, such as mongodb, elasticsearch, mariadb, postgresql, etc. Developers can choose according to the application blockchain.
4. Developer-friendly, for common data applications, it provides a general-purpose data query API. In addition, for complex data, MAN plans to implement Graph Query Language.
5. Controllable indexed data volume, MAN supports multiple data synchronization modes, such as full data synchronization, single application data synchronization, and multiple application combination synchronization. Developers can obtain the data they need through simple configuration.

## Application development process based on MAN

Download the MAN program source code to compile, or directly download the latest MAN-Release program
Modify the relevant configuration file
Run MAN
Use MAN Api and MetaID SDK for development
Carry out development and debugging on the test network of the relevant UTXO chain
Mainnet release

# Build and Run
## Dependencies
### 1. libzmq
The man indexer memory pool data depends on zmq, and libzmq needs to be installed before compilation, otherwise the compilation will fail.

#### OSX
```
brew install zmq
```
#### Linux
Fedora
```
dnf install zeromq-devel
```
#### Ubuntu/Debian/Mint
```
apt-get install libzmq3-dev
```
For more information, please refer to https://zeromq.org/download/

### 2. golang version
MAN's main development language is golang, and go >= 1.20 is required

## Build
```
go mod tidy
go build 
```
## Configuration file
The configuration file is in the same directory as the program running file, and the name must be config.toml
```
[sync]
syncAllData = true  //Whether to save the full amount of pin data
syncBeginTime = ""  //Synchronization data start time
syncEndTime = ""    //Synchronization data end time, empty means synchronization all the time
syncProtocols = ["payLike"] //Specify the specific protocol to be synchronized, which can be multiple
[protocols]
  [protocols.payLike] //Specific protocol data structure
  //Field name
  fields = [{name = "isLike",class = "string",length = 1},
           {name = "likeTo",class = "string",length = 100}]
  //Protocol database index
  indexes = [{fields = ["likeTo"],unique = false},
             {fields = ["pinId"],unique = true},
             {fields = ["pinNumber"],unique = false},
             {fields = ["pinAddress"],unique = false},
            ]
//UTXO chain configuration
[btc]
initialHeight = 2570900 //Default starting block height
//RPC related configuration
rpcHost = "127.0.0.1:18332"
rpcUser = "test"
rpcPass = "test"
rpcHttpPostMode = true
rpcDisableTLS = true
zmqHost = "tcp://127.0.0.1:28336"
//Database configuration
[mongodb]
mongoURI = "mongodb://root:123456@127.0.0.1:27017"
dbName = "man_testnet"
poolSize = 200
timeOut = 20
//Web configuration
[web]
port = ":7777" //Port
//ssl configuration, default to http access if empty
pemFile = ""  //ssl certificate, pem file path
keyFile = ""  //ssl certificate, key file path
Run
./manindexer -chain=btc -databse=mongo -server=1 -test=1
```


# Browser

The MAN indexer comes with a built-in MetaID browser that supports MetaID-related data queries.
This is the deployed and live MAN browser: https://man.metaid.io

## Running
We support the following three deployment and execution methods.
1. ### Compile and run
 - Compile according to the documentation's compilation section.
 - Properly configure the config.toml file in the same directory as the executable.
 -  Run the executable.
```
./manindexer -chain=btc -databse=mongo -server=1 -test=1
```
2. ### Download the release
  - Download the latest release from [here](https://github.com/metaid-developers/man-indexer/releases).
  - Extract the files.
  - Properly configure the config.toml file in the same directory as the executable.
  - Run the executable.
3. ### shell 
  - Download the latest shell file from here.
  - Execute the shell file on the server.
```
./run_manindex.sh
```

To start an HTTP web service, specify the server parameter as 1 when running the program.

`./manindexer -server=1`

## Parameters
```
-chain string
        Which chain to perform indexing (default "btc")
  -database string
        Which database to use (default "mongo")
  -server string
        Run the explorer service (default "1")
  -test string
        Connect to testnet (default "0")
```
The optional values for 'test' are 0, 1, and 2, corresponding to mainnet, testnet3, and regtest networks, respectively.

The service's default port is 80/443. If a specific port needs to be specified, it can be done through the port setting under the web category in the configuration file.

After starting the service, access it via [http://127.0.0.0:{port}](http://127.0.0.0:%7Bport%7D/).

## Browser Features

### Search

Accepts keywords such as MetaID, MetaID Number, Pin Id for querying, but does not support fuzzy search.

### PIN

Lists all PINs in reverse chronological order. Clicking on a PIN allows you to view more information, such as:
[9bc429654d35a11e5dde0136e3466faa03507d7377769743fafa069e38580243i0](https://man.metaid.io/pin/9bc429654d35a11e5dde0136e3466faa03507d7377769743fafa069e38580243i0)

### MetaID

Lists all MetaIDs in descending order of creation time. Clicking on a MetaID allows you to view the PIN that created it.

### Block

Lists all blocks with MetaID protocol data in reverse order of block height. Clicking on a card allows you to view the specific transaction details in that block, such as:
[https://man.metaid.io/block/844453](https://man.metaid.io/block/844453)

### Mempool

Lists MetaID data in the memory pool, which is automatically deleted after being included in a block.

# JSON API

## Basic API

|Endpoint|Method|Parameter|Description|
|---|---|---|---|
|/api/pin/{numberOrId}|GET|PIN number or PIN id|Get PIN details based on PIN number or PIN id|
|/api/address/pin/list/{addressType}/{address}|GET|address, addressType: creator (creator), owner (owner)|Get a list of PINs created or owned by the specified address|
|/api/address/pin/root/{address}|GET|address|Get PIN root based on the specified address|
|/api/node/child/{pinId}|GET|pinId|Get child nodes based on the specified PIN id|
|/api/node/parent/{pinId}|GET|pinId|Get parent nodes based on the specified PIN id|
|/api/info/address/{address}|GET|address|Get MetaID info for the specified address|
|/api/info/rootId/{rootId}|GET|rootId|Get MetaID info based on the specified rootId|
|/api/pin/content/{numberOrId}|GET|PIN number or PIN id|Get the content of a PIN based on PIN number or PIN id|
|/api/getAllPinByParentPath|GET|page, limit, parentPath|Get all PINs based on the specified parentPath|

## generalQuery

General query for protocols data, supports fetching data using get, count, sum methods.

Endpoint: /api/generalQuery

**Method:** POST
```json
{
    "collection": "pins", // Name of the collection to query, required
    "action": "sum", // Query operation, supports get, count, sum 
    "filterRelation": "or", // Query condition relationship, supports or, and (cannot be mixed) 
    "field": [ "number" // Field to return in the query, required for sum operation ], 
     // Query conditions 
    "filter": [
	       { 
	         "operator": "=", // Condition operator, supports =, >, >=, <, <= 
	         "key": "number", // Condition field
	         "value": 1 // Query value 
	       },
	       {
	         "operator": "=", "key": "number", "value": 2
	        }
          ],
	"cursor": 0, // Starting point for returned data 
	"limit": 1, // Number of data records to return
	"sort": [ 
        	"number", // Field to sort by 
		"desc" // Order, supports asc, desc 
        ]
}
```

**Successful Response Example**
```json
{
	"collection": "paylike",
	"action": "get",
	"filterRelation": "and",
	"field": [],
	"filter": [{
		"operator": "=",
		"key": "likeTo",
		"value": "9fec9e5eb879049bd8403ffa45ca0e2756b6c14434b507ccdaf7771d5ec4edf9i0"
	}],
	"cursor": 0,
	"limit": 99999,
	"sort": []
}
```

**Failed Response Example**
```json
{
    "code": -1,
    "message": "Data not found",
    "data": null
}
```
