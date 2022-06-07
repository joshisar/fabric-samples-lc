/*
 * SPDX-License-Identifier: Apache-2.0
 * Source: https://github.com/hyperledger/fabric-samples/releases/tag/v1.4.8
 */

'use strict';

const { FileSystemWallet, Gateway } = require('fabric-network');
const path = require('path');

// EDIT - Changed config path
const ccpPath = path.resolve(
  __dirname,
  '..',
  '..',
  'test-network',
  'organizations',
  'peerOrganizations',
  'org1.example.com',
  'connection-org1.json'
);

async function main() {
  try {
    // Create a new file system based wallet for managing identities.
    const walletPath = path.join(process.cwd(), 'wallet');
    const wallet = new FileSystemWallet(walletPath);
    console.log(`Wallet path: ${walletPath}`);

    // Check to see if we've already enrolled the user.
    const userExists = await wallet.exists('appUser');
    if (!userExists) {
      console.log(
        'An identity for the user "appUser" does not exist in the wallet'
      );
      console.log(
        'Run the registerUser.js application before retrying'
      );
      // EDIT - Return error code when identity is missing
      process.exit(2);
    }

    // Create a new gateway for connecting to our peer node.
    const gateway = new Gateway();
    await gateway.connect(ccpPath, {
      wallet,
      identity: 'appUser',
      discovery: { enabled: true, asLocalhost: true },
    });

    // Get the network (channel) our contract is deployed to.
    const network = await gateway.getNetwork('mychannel');

    // Get the contract from the network.
    const contract = network.getContract('managelc');

    // Evaluate the specified transaction.
    // queryCar transaction - requires 1 argument, ex: ('queryCar', 'CAR4')
    // queryAllCars transaction - requires no arguments, ex: ('queryAllCars')
    const result = await contract.evaluateTransaction(
      'GetLoCById',
      'INLCU0100220002'
    );
    // EDIT - More verbose result print
    console.log(
      'Transaction has been evaluated, result is:',
      JSON.parse(result.toString())
    );
  } catch (error) {
    console.error(`Failed to evaluate transaction: ${error}`);
    process.exit(1);
  }
}

main();
