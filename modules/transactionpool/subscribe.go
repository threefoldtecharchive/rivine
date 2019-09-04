package transactionpool

import (
	"github.com/threefoldtech/rivine/modules"
)

// updateSubscribersTransactions sends a new transaction pool update to all
// subscribers.
func (tp *TransactionPool) updateSubscribersTransactions() error {
	var cc modules.ConsensusChange
	txns := tp.transactionList()
	for _, tSetDiff := range tp.transactionSetDiffs {
		cc = cc.Append(tSetDiff)
	}
	for _, subscriber := range tp.subscribers {
		err := subscriber.ReceiveUpdatedUnconfirmedTransactions(txns, cc)
		if err != nil {
			return err
		}
	}
	return nil
}

// TransactionPoolSubscribe adds a subscriber to the transaction pool.
// Subscribers will receive the full transaction set every time there is a
// significant change to the transaction pool.
func (tp *TransactionPool) TransactionPoolSubscribe(subscriber modules.TransactionPoolSubscriber) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// Add the subscriber to the subscriber list.
	tp.subscribers = append(tp.subscribers, subscriber)

	// Send the new subscriber the transaction pool set.
	txns := tp.transactionList()
	var cc modules.ConsensusChange
	for _, tSetDiff := range tp.transactionSetDiffs {
		cc = cc.Append(tSetDiff)
	}
	subscriber.ReceiveUpdatedUnconfirmedTransactions(txns, cc)
}

// Unsubscribe removes a subscriber from the transaction pool. If the
// subscriber is not in tp.subscribers, Unsubscribe does nothing. If the
// subscriber occurs more than once in tp.subscribers, only the earliest
// occurrence is removed (unsubscription fails).
func (tp *TransactionPool) Unsubscribe(subscriber modules.TransactionPoolSubscriber) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// Search for and remove subscriber from list of subscribers.
	for i := range tp.subscribers {
		if tp.subscribers[i] == subscriber {
			tp.subscribers = append(tp.subscribers[0:i], tp.subscribers[i+1:]...)
			break
		}
	}
}
