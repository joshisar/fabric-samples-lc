package chaincode

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// TODOS:::::::::::::::
// C H A I N C O D E
// checks to ensure , the correct in order movement of states- chaincode, AND accordingly, only correct actions displayed/can-be-taken on UI
// TODO Check in issuance , with same id, present or not, before issuing
// TODO- only advising-org can call its flow
// TODO- only issuing-org can call its flow
// TODO- only negotiating-org can call its flow
// TODO- only parties involved can access read locs from ledger(restrict)

// SmartContract provides functions for creating & managing our LoC
type LocContract struct {
	contractapi.Contract
}

// LoC describes basic details of what makes up a letter of credit
type LoC struct {
	ID                                                  string   `json:"ID"`       // serial number which uniquely identifies the LoC
	DocType                                             string   `json:"doc_type"` // doc_type is used to distinguish the various types of objects in state database
	DocumentaryCreditNumber                             string   `json:"documentary_credit_number"`
	FormOfDocumentaryCredit                             string   `json:"form_of_documentary_credit"`
	DateOfIssue                                         string   `json:"date_of_issue"`
	DateOfExpiry                                        string   `json:"date_of_expiry"`
	PlaceOfExpiry                                       string   `json:"place_of_expiry"`
	ApplicantBank                                       string   `json:"applicant_bank"`
	Applicant                                           string   `json:"applicant"`
	Beneficiary                                         string   `json:"beneficiary"`
	CurrencyCode                                        string   `json:"currency_code"`
	Amount                                              int64    `json:"amount"`
	AvailableWithBy                                     string   `json:"available_with_by"`
	DraftsAt                                            string   `json:"drafts_at"`
	LoadingFrom                                         string   `json:"loading_from"`
	TransportationTo                                    string   `json:"transportation_to"`
	DescriptionOfGoodsAndServices                       string   `json:"description_of_goods_and_services"`
	DocumentsRequired                                   string   `json:"documents_required"`
	Charges                                             string   `json:"charges"`
	PeriodForPresentation                               string   `json:"period_for_presentation"`
	ReimbursingBank                                     string   `json:"reimbursing_bank"`
	InstructionsToThePayingOrAcceptingOrNegotiatingBank string   `json:"instructions_to_the_paying_or_accepting_or_negotiating_bank"`
	AdviseThroughBank                                   string   `json:"advise_through_bank"`
	NegotiatingBank                                     string   `json:"negotiating_bank"`
	IsActive                                            bool     `json:"is_active"` // Is LoC active/expired
	CurrentStatus                                       string   `json:"current_status"`
	StatusLog                                           []string `json:"status_log"`
	DocsUrls                                            []string `json:"docs_urls"`
}

// ********************** HAPPY FLOW START **********************
// ----------------------------------------------------------------
// xxx ISSUANCE_REQUESTED_BY_APPLICANT
// ----------------------------------------------------------------
// vvv ISSUED_BY_APPLICANT_BANK
// vvv ISSUANCE_ACKNOWLEDGED_BY_ADVISING_BANK
// ----------------------------------------------------------------
// vvv AMENDED_BY_APPLICANT_BANK
// vvv AMENDMENT_ACKNOWLEDGED_BY_ADVISING_BANK
// ----------------------------------------------------------------
// vvv AWAITING_DOCUMENTS
// ----------------------------------------------------------------
// xxx DOCUMENTS_SUBMITTED_BY_BENEFICIARY
// xxx PAYMENT_MADE_TO_BENEFICIARY
// ----------------------------------------------------------------
// vvv DOCUMENTS_SUBMITTED_BY_NEGOTIATING_BANK
// vvv DOCUMENTS_ACCEPTED_BY_APPLICANT_BANK
// ----------------------------------------------------------------
// vvv PAYMENT_MADE_TO_NEGOTIATING_BANK
// ----------------------------------------------------------------
// vvv PAYMENT_ACKNOWLEDGED_BY_NEGOTIATING_BANK
// ----------------------------------------------------------------
// vvv CLOSED_BY_APPLICANT_BANK
// ----------------------------------------------------------------
// ********************** HAPPY FLOW END **********************

// -------------------------------------------------------------------------------------------------------------------------------------
// IssueLoC issues a new LoC and puts on the ledger
func (c *LocContract) IssueLoC(ctx contractapi.TransactionContextInterface, jsonLoC string) (*LoC, error) {
	// only applicant bank can do it- check
	// Un-Marshal jsonLoC to loc
	var loc LoC
	json.Unmarshal([]byte(jsonLoC), &loc)
	// current status
	loc.CurrentStatus = "ISSUED_BY_APPLICANT_BANK"
	// status log
	current_time := GetTodaysDateTimeFormatted()
	status := fmt.Sprintf("LoC issued by %s on %s", loc.ApplicantBank, current_time)
	loc.StatusLog = append(loc.StatusLog, status)
	// is_Active
	loc.IsActive = true
	// doc_urls - empty array of strings initialised on its own
	loc.DocsUrls = make([]string, 0)
	// Marshal loc
	locJSON, err := json.Marshal(loc)
	if err != nil {
		log.Println("error -> json.Marshal -> IssueLoC\n", err)
		return nil, fmt.Errorf("failed to marshal into Json: %v", err)
	}
	// Put on ledger
	err = ctx.GetStub().PutState(loc.ID, locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.PutState -> IssueLoC\n", err)
		return nil, fmt.Errorf("failed to put on ledger: %v", err)
	}
	// Emit the LoCIssued event
	err = ctx.GetStub().SetEvent("LoCIssued", locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.SetEvent -> IssueLoC\n", err)
		return nil, fmt.Errorf("failed to set event: %v", err)
	}
	return &loc, nil
}

// -------------------------------------------------------------------------------------------------------------------------------------
// AcknowledgeLoCIssuance acknowledges issued LoC and updates status
func (c *LocContract) AcknowledgeLoCIssuance(ctx contractapi.TransactionContextInterface, id string) (*LoC, error) {
	// only advising bank can do it- check
	// Get LoC if exists
	loc, err := c.GetLoCById(ctx, id)
	if err != nil {
		log.Println("error -> c.GetLoCById -> AcknowledgeLoCIssuance\n", err)
		return nil, fmt.Errorf("LoC with Id@%s does not exist", id)
	}
	// current status
	loc.CurrentStatus = "ISSUANCE_ACKNOWLEDGED_BY_ADVISING_BANK"
	// status log
	current_time := GetTodaysDateTimeFormatted()
	status := fmt.Sprintf("LoC issuance acknowledged by %s on %s", loc.AdviseThroughBank, current_time)
	loc.StatusLog = append(loc.StatusLog, status)
	// Marshal loc
	locJSON, err := json.Marshal(loc)
	if err != nil {
		log.Println("error -> json.Marshal -> AcknowledgeLoCIssuance\n", err)
		return nil, fmt.Errorf("failed to marshal into Json: %v", err)
	}
	// Put on ledger
	err = ctx.GetStub().PutState(loc.ID, locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.PutState -> AcknowledgeLoCIssuance\n", err)
		return nil, fmt.Errorf("failed to put on ledger: %v", err)
	}
	// Emit the LoCIssuanceAcknowledged event
	err = ctx.GetStub().SetEvent("LoCIssuanceAcknowledged", locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.SetEvent -> AcknowledgeLoCIssuance\n", err)
		return nil, fmt.Errorf("failed to set event: %v", err)
	}
	return loc, nil
}

// -------------------------------------------------------------------------------------------------------------------------------------
// AmendLoCAmount amends the LoC amount for LoC with given {id} and amount
func (c *LocContract) AmendLoCAmount(ctx contractapi.TransactionContextInterface, id string, amount int64) (*LoC, error) {
	// only applicant bank can do it- check
	// Get LoC if exists
	loc, err := c.GetLoCById(ctx, id)
	if err != nil {
		log.Println("error -> c.GetLoCById -> AmendLoCAmount\n", err)
		return nil, fmt.Errorf("LoC with Id@%s does not exist", id)
	}
	// Update LoC amount
	oldAmount := loc.Amount
	loc.Amount = amount
	// current status
	loc.CurrentStatus = "AMENDED_BY_APPLICANT_BANK"
	// status log
	current_time := GetTodaysDateTimeFormatted()
	status := fmt.Sprintf("LoC amount amednded by %s from %d to %d on %s", loc.ApplicantBank, oldAmount, amount, current_time)
	loc.StatusLog = append(loc.StatusLog, status)
	// Marshal loc
	locJSON, err := json.Marshal(loc)
	if err != nil {
		log.Println("error -> json.Marshal -> AmendLoCAmount\n", err)
		return nil, fmt.Errorf("failed to marshal into Json: %v", err)
	}
	// Put on ledger
	err = ctx.GetStub().PutState(loc.ID, locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.PutState -> AmendLoCAmount\n", err)
		return nil, fmt.Errorf("failed to put on ledger: %v", err)
	}
	// Emit the LoCAmountAmended event
	err = ctx.GetStub().SetEvent("LoCAmountAmended", locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.SetEvent -> AmendLoCAmount\n", err)
		return nil, fmt.Errorf("failed to set event: %v", err)
	}
	return loc, nil
}

// -------------------------------------------------------------------------------------------------------------------------------------
// AcknowledgeLoCAmendment acknowledges amended LoC and updates status
func (c *LocContract) AcknowledgeLoCAmendment(ctx contractapi.TransactionContextInterface, id string) (*LoC, error) {
	// only advising bank can do it- check
	// Get LoC if exists
	loc, err := c.GetLoCById(ctx, id)
	if err != nil {
		log.Println("error -> c.GetLoCById -> AcknowledgeLoCAmendment\n", err)
		return nil, fmt.Errorf("LoC with Id@%s does not exist", id)
	}
	// current status
	loc.CurrentStatus = "AMENDMENT_ACKNOWLEDGED_BY_ADVISING_BANK"
	// status log
	current_time := GetTodaysDateTimeFormatted()
	status := fmt.Sprintf("LoC amendment acknowledged by %s on %s", loc.AdviseThroughBank, current_time)
	loc.StatusLog = append(loc.StatusLog, status)
	// current status
	loc.CurrentStatus = "AWAITING_DOCUMENTS"
	// status log
	status = fmt.Sprintf("%s awaiting documents from %s", loc.ApplicantBank, loc.NegotiatingBank)
	loc.StatusLog = append(loc.StatusLog, status)
	// Marshal loc
	locJSON, err := json.Marshal(loc)
	if err != nil {
		log.Println("error -> json.Marshal -> AcknowledgeLoCAmendment\n", err)
		return nil, fmt.Errorf("failed to marshal into Json: %v", err)
	}
	// Put on ledger
	err = ctx.GetStub().PutState(loc.ID, locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.PutState -> AcknowledgeLoCAmendment\n", err)
		return nil, fmt.Errorf("failed to put on ledger: %v", err)
	}
	// Emit the LoCAmendmentAcknowledged event
	err = ctx.GetStub().SetEvent("LoCAmendmentAcknowledged", locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.SetEvent -> AcknowledgeLoCAmendment\n", err)
		return nil, fmt.Errorf("failed to set event: %v", err)
	}
	return loc, nil
}

// -------------------------------------------------------------------------------------------------------------------------------------
// SubmitDocuments just updates in ledger that docs for given LoC {id} has been submitted, TODO S3
func (c *LocContract) SubmitDocuments(ctx contractapi.TransactionContextInterface, id string, jsonDocsUrls string) (*LoC, error) {
	// only negotiating bank can do it- check
	docsUrls := []string{}
	json.Unmarshal([]byte(jsonDocsUrls), &docsUrls)
	log.Println("docsUrls >>>>>> \n", docsUrls)
	// Get LoC if exists
	loc, err := c.GetLoCById(ctx, id)
	if err != nil {
		log.Println("error -> c.GetLoCById -> SubmitDocuments\n", err)
		return nil, fmt.Errorf("LoC with Id@%s does not exist", id)
	}
	// append to docs Urls array
	loc.DocsUrls = docsUrls
	// current status
	loc.CurrentStatus = "DOCUMENTS_SUBMITTED_BY_NEGOTIATING_BANK"
	// status log
	current_time := GetTodaysDateTimeFormatted()
	status := fmt.Sprintf("Document(s) submitted by %s to %s on %s", loc.NegotiatingBank, loc.ApplicantBank, current_time)
	loc.StatusLog = append(loc.StatusLog, status)
	// Marshal loc
	locJSON, err := json.Marshal(loc)
	if err != nil {
		log.Println("error -> json.Marshal -> SubmitDocuments\n", err)
		return nil, fmt.Errorf("failed to marshal into Json: %v", err)
	}
	// Put on ledger
	err = ctx.GetStub().PutState(loc.ID, locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.PutState -> SubmitDocuments\n", err)
		return nil, fmt.Errorf("failed to put on ledger: %v", err)
	}
	// Emit the DocumentsSubmitted event
	err = ctx.GetStub().SetEvent("DocumentsSubmitted", locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.SetEvent -> SubmitDocuments\n", err)
		return nil, fmt.Errorf("failed to set event: %v", err)
	}
	return loc, nil
}

// -------------------------------------------------------------------------------------------------------------------------------------
// AcceptDocuments is done after its accepted by applicant bank for given LoC, it updates status
func (c *LocContract) AcceptDocuments(ctx contractapi.TransactionContextInterface, id string) (*LoC, error) {
	// only applicant bank can do it- check
	// Get LoC if exists
	loc, err := c.GetLoCById(ctx, id)
	if err != nil {
		log.Println("error -> c.GetLoCById -> AcceptDocuments\n", err)
		return nil, fmt.Errorf("LoC with Id@%s does not exist", id)
	}
	// current status
	loc.CurrentStatus = "DOCUMENTS_ACCEPTED_BY_APPLICANT_BANK"
	// status log
	current_time := GetTodaysDateTimeFormatted()
	status := fmt.Sprintf("Documents accepted by %s from %s on %s", loc.ApplicantBank, loc.NegotiatingBank, current_time)
	loc.StatusLog = append(loc.StatusLog, status)
	// Marshal loc
	locJSON, err := json.Marshal(loc)
	if err != nil {
		log.Println("error -> json.Marshal -> AcceptDocuments\n", err)
		return nil, fmt.Errorf("failed to marshal into Json: %v", err)
	}
	// Put on ledger
	err = ctx.GetStub().PutState(loc.ID, locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.PutState -> AcceptDocuments\n", err)
		return nil, fmt.Errorf("failed to put on ledger: %v", err)
	}
	// Emit the LoCAmendmentAcknowledged event
	err = ctx.GetStub().SetEvent("DocumentsAccepted", locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.SetEvent -> AcceptDocuments\n", err)
		return nil, fmt.Errorf("failed to set event: %v", err)
	}
	return loc, nil
}

// -------------------------------------------------------------------------------------------------------------------------------------
// ConfirmPayment just updates in ledger that payment to negotiating bank for given LoC {id} has been done, TODO S3
func (c *LocContract) ConfirmPayment(ctx contractapi.TransactionContextInterface, id string) (*LoC, error) {
	// only applicant bank can do it- check
	// Get LoC if exists
	loc, err := c.GetLoCById(ctx, id)
	if err != nil {
		log.Println("error -> c.GetLoCById -> ConfirmPayment\n", err)
		return nil, fmt.Errorf("LoC with Id@%s does not exist", id)
	}
	// current status
	loc.CurrentStatus = "PAYMENT_DONE_FROM_APPLICANT_BANK_TO_NEGOTIATING_BANK"
	// status log
	current_time := GetTodaysDateTimeFormatted()
	status := fmt.Sprintf("Payment confirmed from %s to %s on %s", loc.ApplicantBank, loc.NegotiatingBank, current_time)
	loc.StatusLog = append(loc.StatusLog, status)
	// Marshal loc
	locJSON, err := json.Marshal(loc)
	if err != nil {
		log.Println("error -> json.Marshal -> ConfirmPayment\n", err)
		return nil, fmt.Errorf("failed to marshal into Json: %v", err)
	}
	// Put on ledger
	err = ctx.GetStub().PutState(loc.ID, locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.PutState -> ConfirmPayment\n", err)
		return nil, fmt.Errorf("failed to put on ledger: %v", err)
	}
	// Emit the PaymentConfirmed event
	err = ctx.GetStub().SetEvent("PaymentConfirmed", locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.SetEvent -> ConfirmPayment\n", err)
		return nil, fmt.Errorf("failed to set event: %v", err)
	}
	return loc, nil
}

// -------------------------------------------------------------------------------------------------------------------------------------
// AcknowledgePayment is done after payment_receive is checked by negotiating bank for given LoC, it updates status
func (c *LocContract) AcknowledgePayment(ctx contractapi.TransactionContextInterface, id string) (*LoC, error) {
	// only negotiating bank can do it- check
	// Get LoC if exists
	loc, err := c.GetLoCById(ctx, id)
	if err != nil {
		log.Println("error -> c.GetLoCById -> AcknowledgePayment\n", err)
		return nil, fmt.Errorf("LoC with Id@%s does not exist", id)
	}
	// current status
	loc.CurrentStatus = "PAYMENT_ACKNOWLEDGED_FROM_APPLICANT_BANK_TO_NEGOTIATING_BANK"
	// status log
	current_time := GetTodaysDateTimeFormatted()
	status := fmt.Sprintf("Payment acknowledged from %s to %s on %s", loc.ApplicantBank, loc.NegotiatingBank, current_time)
	loc.StatusLog = append(loc.StatusLog, status)
	// Marshal loc
	locJSON, err := json.Marshal(loc)
	if err != nil {
		log.Println("error -> json.Marshal -> AcknowledgePayment\n", err)
		return nil, fmt.Errorf("failed to marshal into Json: %v", err)
	}
	// Put on ledger
	err = ctx.GetStub().PutState(loc.ID, locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.PutState -> AcknowledgePayment\n", err)
		return nil, fmt.Errorf("failed to put on ledger: %v", err)
	}
	// Emit the PaymentAcknowledged event
	err = ctx.GetStub().SetEvent("PaymentAcknowledged", locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.SetEvent -> AcknowledgePayment\n", err)
		return nil, fmt.Errorf("failed to set event: %v", err)
	}
	return loc, nil
}

// -------------------------------------------------------------------------------------------------------------------------------------
// CloseLoC closes the LoC with given {id}
func (c *LocContract) CloseLoC(ctx contractapi.TransactionContextInterface, id string) (*LoC, error) {
	// only applicant bank can do it- check
	// Get LoC Json if exists
	loc, err := c.GetLoCById(ctx, id)
	if err != nil {
		log.Println("error -> c.GetLoCById -> CloseLoC\n", err)
		return nil, fmt.Errorf("LoC with Id@%s does not exist", id)
	}
	// current status
	loc.CurrentStatus = "CLOSED_BY_APPLICANT_BANK"
	// status log
	current_time := GetTodaysDateTimeFormatted()
	status := fmt.Sprintf("LoC closed by %s on %s", loc.ApplicantBank, current_time)
	loc.StatusLog = append(loc.StatusLog, status)
	// mark as inactive
	loc.IsActive = false
	// Marshal loc
	locJSON, err := json.Marshal(loc)
	if err != nil {
		log.Println("error -> json.Marshal -> CloseLoC\n", err)
		return nil, fmt.Errorf("failed to marshal into Json: %v", err)
	}
	// Put on ledger
	err = ctx.GetStub().PutState(loc.ID, locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.PutState -> CloseLoC\n", err)
		return nil, fmt.Errorf("failed to put on ledger: %v", err)
	}
	// Emit the LoCClosed event
	err = ctx.GetStub().SetEvent("LoCClosed", locJSON)
	if err != nil {
		log.Println("error -> ctx.GetStub.SetEvent -> CloseLoC\n", err)
		return nil, fmt.Errorf("failed to set event: %v", err)
	}
	return loc, nil
}

// -------------------------------------------------------------------------------------------------------------------------------------
// GetLoCById returns the LoC stored in the channel with given {id}
func (c *LocContract) GetLoCById(ctx contractapi.TransactionContextInterface, id string) (*LoC, error) {
	// Get LoC Json if exists
	locJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		log.Println("error -> ctx.GetStub.GetState -> GetLoCById\n", err)
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if locJSON == nil {
		log.Printf("the LoC with Id@%s does not exist\n", id)
		return nil, fmt.Errorf("the LoC with Id@%s does not exist", id)
	}
	// Unmarshal and return LoC
	var loc LoC
	err = json.Unmarshal(locJSON, &loc)
	if err != nil {
		log.Println("error -> json.Unmarshal -> GetLoCById\n", err)
		return nil, fmt.Errorf("failed to unmarshal from Json: %v", err)
	}
	return &loc, nil
}

// -------------------------------------------------------------------------------------------------------------------------------------
// GetIssuedLoCs returns issued LCs for org of invoking client
func (c *LocContract) GetIssuedLoCs(ctx contractapi.TransactionContextInterface) ([]*LoC, error) {
	org, _ := getOrgName(ctx)
	// log.Println("org", org)
	// Query string
	// queryString := fmt.Sprintf(`{"selector":{"docType":"asset","owner":"%s"}}`, owner)
	queryString := fmt.Sprintf(`{"selector":{"doc_type":"LoC","applicant_bank":"%s"}}`, org)
	log.Println("queryString", queryString)
	// Get result iterator
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		log.Println("error -> ctx.GetStub.GetQueryResult -> GetIssuedLoCs\n", err)
		return nil, fmt.Errorf("failed to get query result: %v", err)
	}
	defer resultsIterator.Close()
	var locs []*LoC
	// Iterate the results, unmarshal & append to locs & return
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			log.Println("error -> resultsIterator.Next -> GetIssuedLoCs\n", err)
			return nil, fmt.Errorf("failed to read from result iterator: %v", err)
		}
		var loc LoC
		err = json.Unmarshal(queryResult.Value, &loc)
		if err != nil {
			log.Println("error -> json.Unmarshal -> GetIssuedLoCs\n", err)
			return nil, fmt.Errorf("failed to unmarshal query result: %v", err)
		}
		locs = append(locs, &loc)
	}
	return locs, nil
}

// -------------------------------------------------------------------------------------------------------------------------------------
// GetAdvisedLoCs returns advising LCs for org of invoking client
func (c *LocContract) GetAdvisingLoCs(ctx contractapi.TransactionContextInterface) ([]*LoC, error) {
	org, _ := getOrgName(ctx)
	// Query string
	queryString := fmt.Sprintf(`{"selector":{"doc_type":"LoC","advise_through_bank":"%s"}}`, org)
	// Get result iterator
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		log.Println("error -> ctx.GetStub.GetQueryResult -> GetAdvisingLoCs\n", err)
		return nil, fmt.Errorf("failed to get query result: %v", err)
	}
	defer resultsIterator.Close()
	var locs []*LoC
	// Iterate the results, unmarshal & append to locs & return
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			log.Println("error -> resultsIterator.Next -> GetAdvisingLoCs\n", err)
			return nil, fmt.Errorf("failed to read from result iterator: %v", err)
		}
		var loc LoC
		err = json.Unmarshal(queryResult.Value, &loc)
		if err != nil {
			log.Println("error -> json.Unmarshal -> GetAdvisingLoCs\n", err)
			return nil, fmt.Errorf("failed to unmarshal query result: %v", err)
		}
		locs = append(locs, &loc)
	}
	return locs, nil
}

// -------------------------------------------------------------------------------------------------------------------------------------
// GetNegotiatingLoCs returns negotiating LCs for org of invoking client
func (c *LocContract) GetNegotiatingLoCs(ctx contractapi.TransactionContextInterface) ([]*LoC, error) {
	org, _ := getOrgName(ctx)
	// Query string
	queryString := fmt.Sprintf(`{"selector":{"doc_type":"LoC","negotiating_bank":"%s"}}`, org)
	// Get result iterator
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		log.Println("error -> ctx.GetStub.GetQueryResult -> GetNegotiatingLoCs\n", err)
		return nil, fmt.Errorf("failed to get query result: %v", err)
	}
	defer resultsIterator.Close()
	var locs []*LoC
	// Iterate the results, unmarshal & append to locs & return
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			log.Println("error -> resultsIterator.Next -> GetNegotiatingLoCs\n", err)
			return nil, fmt.Errorf("failed to read from result iterator: %v", err)
		}
		var loc LoC
		err = json.Unmarshal(queryResult.Value, &loc)
		if err != nil {
			log.Println("error -> json.Unmarshal -> GetNegotiatingLoCs\n", err)
			return nil, fmt.Errorf("failed to unmarshal query result: %v", err)
		}
		locs = append(locs, &loc)
	}
	return locs, nil
}

// -------------------------------------------------------------------------------------------------------------------------------------
// InitLedger
func (c *LocContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// creating hard-coded first LC- test
	locs := []*LoC{{ID: "INLCU0100220001", DocType: "LoC", DocumentaryCreditNumber: "INLCU0100220001", FormOfDocumentaryCredit: "IRREVOCABLE", DateOfIssue: "20220105", DateOfExpiry: "20220221", PlaceOfExpiry: "NEGOTIATION BANK COUNTER", ApplicantBank: "Org1", Applicant: "AMBER ENTERPRISES INDIA LTD, C-3, SITE-IV, UPSIDC IND. AREA, KASNA ROAD, GREATER NOIDA-201305, U.P, INDIA", Beneficiary: "POSCO INDIA PROCESSING CENTER PVT", CurrencyCode: "INR", Amount: 11436300, AvailableWithBy: "ANY BANK IN INDIA BY NEGOTIATION", DraftsAt: "90 DAYS FROM THE DATE OF BILL OF EXCHANGE", LoadingFrom: "ANYWHERE IN INDIA", TransportationTo: "ANYWHERE IN INDIA", DescriptionOfGoodsAndServices: "100 MT OF GI SHEET AS PER PI NO. POSCO-IHPL/PI/AEPL/JAN2022/01 DTD 04.01.2022, HS CODE:72104900, CIP, ANY WHERE IN INDIA, INCOTERMS 2020", DocumentsRequired: "1: BILL OF EXCHANGE WILL BE PRESENTED AFTER DEDUCTION OF TDS AT 0.1 PCT ON BASIC VALUE OF THE INVOICE. 2: TAX INVOICE IN ONE ORIGINAL. 3: ORIGINAL LORRY RECEIPT ISSUED BY NON IBA APPROVED TRANSPORTER CONSIGNED TO RBL BANK LTD NOTIFY APPLICANT AND MARKED FREIGHT PREPAID. 4.INSURANCE POLICY/CERTIFICATE IN THE CURRENCY OF THE CREDIT AND BLANK ENDORSED FOR CIP VALUE OF GOODS PLUS 10 PCT SHOWING CLAIMS PAYABLE IN INDIA IRRESPECTIVE OF PERCENTAGE. 5: INSURANCE TO COVER ALL RISKS FROM SUPPLIER WAREHOUSE TO APPLICANT WAREHOUSE.", Charges: "APPLICANT BANK CHARGES TO APPLICANT ACCOUNT AND BENEFICIARY ACCOUNT INCLUDING DISCREPANCY CHARGES TO BENEFICIARY ACCOUNT", PeriodForPresentation: "WITHIN 21 DAYS FROM THE DATE OF SHIPMENT BUT WITHIN THE VALIDITY OF THE LC.", ReimbursingBank: "Org1", InstructionsToThePayingOrAcceptingOrNegotiatingBank: "UPON SUBMISSION OF CREDIT COMPLIANT DOCUMENTS, WE WILL REIMBURSE YOU ON DUE DATE AS PER YOUR INSTRUCTIONS", AdviseThroughBank: "Org2", NegotiatingBank: "Org2", IsActive: true, CurrentStatus: "ISSUED_BY_APPLICANT_BANK", StatusLog: []string{"LoC issued by Org1 on Apr 11, 2022 at 11:46 AM"}, DocsUrls: []string{"https://bafybeidbwaneqilaaytdvwspd6f4mvashv6wbguqxsbawbp23sbz4ypjcy.ipfs.infura-ipfs.io"}}}
	for _, loc := range locs {
		locJSON, err := json.Marshal(loc)
		if err != nil {
			log.Println("error -> json.Marshal -> InitLedger\n", err)
		}
		err = ctx.GetStub().PutState(loc.ID, locJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}
	return nil
}
