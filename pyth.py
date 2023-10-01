"""
no imports
just a proof of concept design
"""

w3 = Web3(Web3.AsyncHTTPProvider("http://127.0.0.1:8545"), modules={'eth': (AsyncEth,)}, middlewares=[])

start_block = 12000000
end_block__ = 14000000 



transactions = {}

## coRoutine 1, we can probably lookup for 10 blocks or more at same time
async def getBlock(b):

	print('now getting all transactions from block number', b)

	gb = w3.eth.get_block(b)

	txs = gb['transactions'] 

	for tx in txs:
		asyncio.ensure_future(getTx(tx))



## a block typically has between 50 and 200 transactions
## we could have more concurrent workers working here to retrieve the details quickly
async def getTx(tx):

	print('now getting details from transaction ID ', tx)
	
	gt = w3.eth.get_transaction(tx)

	txValue = gt['value']

	if txValue > 0:
		transactions[tx] = {
		'txData': gt['data'],
		'txFrom': gt['from'],
		'txTo': gt['to']
		}


async def amain():
	for b in range(start_block,end_block__):
		asyncio.ensure_future(getBlock(b))


if __name__ == '__main__':

	startLoop = amain()

	


