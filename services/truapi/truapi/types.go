package truapi

import (
	"time"

	app "github.com/TruStory/truchain/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tcmn "github.com/tendermint/tendermint/libs/common"
)

// CredArgument represents an argument that earned cred based on likes.
type CredArgument struct {
	ID        int64          `json:"id" graphql:"id" `
	StoryID   int64          `json:"storyId" graphql:"storyId"`
	Body      string         `json:"body"`
	Creator   sdk.AccAddress `json:"creator" `
	Timestamp app.Timestamp  `json:"timestamp"`
	Vote      bool           `json:"vote"`
	Amount    sdk.Coin       `json:"coin"`
}

// CommentNotificationRequest is the payload sent to pushd for sending notifications.
type CommentNotificationRequest struct {
	// ID is the comment id.
	ID              int64     `json:"id"`
	ArgumentCreator string    `json:"argument_creator"`
	ArgumentID      int64     `json:"argumentId"`
	StoryID         int64     `json:"storyId"`
	Creator         string    `json:"creator"`
	Timestamp       time.Time `json:"timestamp"`
}

// V2 Truchain structs

// AppAccount will be imported from truchain in the future
type AppAccount struct {
	BaseAccount

	EarnedStake []EarnedCoin
	SlashCount  int
	IsJailed    bool
	JailEndTime time.Time
	CreatedTime time.Time
}

// EarnedCoin will be imported from truchain in the future
type EarnedCoin struct {
	sdk.Coin

	CommunityID int64
}

// BaseAccount will be imported from truchain in the future
type BaseAccount struct {
	Address       string
	Coins         sdk.Coins
	PubKey        tcmn.HexBytes
	AccountNumber uint64
	Sequence      uint64
}
