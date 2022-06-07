/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { FileSystemWallet, Gateway } = require('fabric-network');
const fs = require('fs');
const path = require('path');

// load the network configuration
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

    // Submit the specified transaction.
    const LC = {
      ID: 'INLCU0100220002',
      doc_type: 'LoC',
      documentary_credit_number: 'INLCU0100220002',
      form_of_documentary_credit: 'IRREVOCABLE',
      date_of_issue: '20220105',
      date_of_expiry: '20220221',
      place_of_expiry: 'NEGOTIATION BANK COUNTER',
      applicant_bank: 'Org1',
      applicant:
        'AMBER ENTERPRISES INDIA LTD, C-3, SITE-IV, UPSIDC IND. AREA, KASNA ROAD, GREATER NOIDA-201305, U.P, INDIA',
      beneficiary: 'POSCO INDIA PROCESSING CENTER PVT',
      currency_code: 'INR',
      amount: 1200000,
      available_with_by: 'ANY BANK IN INDIA BY NEGOTIATION',
      drafts_at: '90 DAYS FROM THE DATE OF BILL OF EXCHANGE',
      loading_from: 'ANYWHERE IN INDIA',
      transportation_to: 'ANYWHERE IN INDIA',
      description_of_goods_and_services:
        '100 MT OF GI SHEET AS PER PI NO. POSCO-IHPL/PI/AEPL/JAN2022/01 DTD 04.01.2022, HS CODE:72104900, CIP, ANY WHERE IN INDIA, INCOTERMS 2020',
      documents_required:
        '1: BILL OF EXCHANGE WILL BE PRESENTED AFTER DEDUCTION OF TDS AT 0.1 PCT ON BASIC VALUE OF THE INVOICE. 2: TAX INVOICE IN ONE ORIGINAL. 3: ORIGINAL LORRY RECEIPT ISSUED BY NON IBA APPROVED TRANSPORTER CONSIGNED TO RBL BANK LTD NOTIFY APPLICANT AND MARKED FREIGHT PREPAID. 4.INSURANCE POLICY/CERTIFICATE IN THE CURRENCY OF THE CREDIT AND BLANK ENDORSED FOR CIP VALUE OF GOODS PLUS 10 PCT SHOWING CLAIMS PAYABLE IN INDIA IRRESPECTIVE OF PERCENTAGE. 5: INSURANCE TO COVER ALL RISKS FROM SUPPLIER WAREHOUSE TO APPLICANT WAREHOUSE.',
      charges:
        'APPLICANT BANK CHARGES TO APPLICANT ACCOUNT AND BENEFICIARY ACCOUNT INCLUDING DISCREPANCY CHARGES TO BENEFICIARY ACCOUNT',
      period_for_presentation:
        'WITHIN 21 DAYS FROM THE DATE OF SHIPMENT BUT WITHIN THE VALIDITY OF THE LC.',
      reimbursing_bank: 'Org1',
      instructions_to_the_paying_or_accepting_or_negotiating_bank:
        'UPON SUBMISSION OF CREDIT COMPLIANT DOCUMENTS, WE WILL REIMBURSE YOU ON DUE DATE AS PER YOUR INSTRUCTIONS',
      advise_through_bank: 'Org2',
      negotiating_bank: 'Org2',
      is_active: true,
      current_status: 'ISSUED_BY_APPLICANT_BANK',
      status_log: ['LoC issued by Org1 on Apr 11, 2022 at 11:46 AM'],
      docs_urls: [
        'https://bafybeidbwaneqilaaytdvwspd6f4mvashv6wbguqxsbawbp23sbz4ypjcy.ipfs.infura-ipfs.io',
      ],
    };
    const LCJson = JSON.stringify(LC);
    const result = await contract.submitTransaction(
      'IssueLoC',
      LCJson
    );
    console.log(`Transaction Committed *** RESULT *** : ${result}`);

    // Disconnect from the gateway.
    await gateway.disconnect();
  } catch (error) {
    console.error(`Failed to submit transaction: ${error}`);
    process.exit(1);
  }
}

main();
