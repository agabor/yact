package logic

func LoadContextForMessageType(transactionType TransactionType) ([]Transaction, error) {
	transactions, err := LoadContext()
	if err != nil {
		return nil, err
	}

	newTransactions := make([]Transaction, 0)
	newTransaction := Transaction{Type: transactionType}

	allowedTypes := make(map[TransactionType]bool)

	switch transactionType {
	case TransactionTypeQuestion:
		allowedTypes[TransactionTypeQuestion] = true
		allowedTypes[TransactionTypePlan] = true
	case TransactionTypePlan:
		allowedTypes[TransactionTypeQuestion] = true
		allowedTypes[TransactionTypePlan] = true
	case TransactionTypeAct:
		allowedTypes[TransactionTypeAct] = true
	}

	lastPlan := ""
	for _, tx := range transactions {
		if allowedTypes[tx.Type] {
			newTransactions = append(newTransactions, tx)
		} else {
			for _, file := range tx.Context {
				newTransaction.Context = append(newTransaction.Context, file)
			}
		}
		if tx.Type == TransactionTypePlan {
			lastPlan = tx.Response
		}
	}

	if transactionType == TransactionTypeAct {
		newTransaction.Request = []string{lastPlan}
	}

	return append(newTransactions, newTransaction), nil
}

type MessageType string

const (
	MessageTypeFile      MessageType = "File"
	MessageTypeQuestion  MessageType = "Question"
	MessageTypeAnswer    MessageType = "Answer"
	MessageTypeCommand   MessageType = "Command"
	MessageTypeAction    MessageType = "Action"
	MessageTypeObjective MessageType = "Objective"
	MessageTypePlan      MessageType = "Plan"
	MessageTypeRevision  MessageType = "Revision"
)

func ResponseType(messageType MessageType) MessageType {
	switch messageType {
	case MessageTypeCommand:
		return MessageTypeAction
	case MessageTypeQuestion:
		return MessageTypeAnswer
	case MessageTypeObjective:
		return MessageTypePlan
	default:
		return messageType
	}
}

type Message struct {
	Type    MessageType
	Path    string
	Content string
}

type TransactionType string

const (
	TransactionTypeNone     TransactionType = "None"
	TransactionTypeQuestion TransactionType = "Question"
	TransactionTypeAct      TransactionType = "Act"
	TransactionTypePlan     TransactionType = "Plan"
)

type File struct {
	Path    string
	Content string
}

type Transaction struct {
	Type     TransactionType
	Request  []string
	Response string
	Context  []File
}
