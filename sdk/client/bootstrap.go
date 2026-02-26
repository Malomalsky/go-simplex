package client

import (
	"context"
	"fmt"

	"github.com/Malomalsky/go-simplex/sdk/types"
)

type BootstrapResult struct {
	User    *types.User
	Address string
}

func (c *Client) BootstrapBot(ctx context.Context) (*BootstrapResult, error) {
	user, err := c.GetActiveUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("get active user: %w", err)
	}

	address, err := c.EnsureUserAddress(ctx, user.UserID)
	if err != nil {
		return nil, fmt.Errorf("ensure user address: %w", err)
	}

	if err := c.EnableAddressAutoAccept(ctx, user.UserID); err != nil {
		return nil, fmt.Errorf("enable address auto-accept: %w", err)
	}

	return &BootstrapResult{
		User:    user,
		Address: address,
	}, nil
}
