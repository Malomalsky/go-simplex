# go-simplex

Go SDK for building bots on top of the official SimpleX Bot API.

Current stage: implementation in progress.

## Documents

- API research: `docs/research/upstream-api.md`
- TypeScript SDK research: `docs/research/upstream-sdk.md`
- Implementation roadmap: `docs/plan/go-sdk-roadmap.md`

## Principles

- full parity with documented SimpleX bot API
- idiomatic Go API
- generated contracts to prevent upstream drift
- practical bot-developer ergonomics

## Development

Generate command catalog from upstream snapshot:

```bash
go run ./cmd/simplexgen
```

Run tests:

```bash
go test ./...
```

Smoke check against running SimpleX CLI websocket:

```bash
go run ./cmd/simplex-smoke --ws ws://localhost:5225
```

## Quickstart (current API)

Run SimpleX CLI with websocket API:

```bash
simplex-chat -p 5225
```

Create client and bot runtime:

```go
ctx := context.Background()
cli, err := client.NewWebSocket(ctx, "ws://localhost:5225")
if err != nil {
    panic(err)
}
defer cli.Close(ctx)

rt, err := bot.NewRuntime(cli)
if err != nil {
    panic(err)
}

bot.OnDirectText(rt, func(ctx context.Context, cli *client.Client, msg bot.DirectTextMessage) error {
    return msg.Reply(ctx, cli, "echo: "+msg.Text)
})

if err := rt.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
    panic(err)
}
```

Bot runtime helpers:

- `bot.OnTyped`
- `bot.OnDirectText`
- `bot.ExtractDirectTextMessages`

High-level helper methods currently available on `*client.Client`:

- `BootstrapBot`
- `GetActiveUser`
- `GetUserAddress`
- `CreateUserAddress`
- `DeleteUserAddress`
- `SetProfileAddress`
- `SetAddressSettings`
- `EnsureUserAddress`
- `ListContacts`
- `ListGroups`
- `CreateContactInvitation`
- `ConnectPlan`
- `ConnectWithPreparedLink`
- `ConnectWithLink`
- `AcceptContactRequest`
- `RejectContactRequest`
- `CreateUser`
- `ListUsers`
- `SetActiveUser`
- `DeleteUser`
- `UpdateProfile`
- `SetContactPreferences`
- `AddGroupMember`
- `JoinGroup`
- `AcceptGroupMember`
- `SetGroupMembersRole`
- `BlockGroupMembersForAll`
- `RemoveGroupMembers`
- `LeaveGroup`
- `ListGroupMembers`
- `CreateGroup`
- `UpdateGroupProfile`
- `CreateGroupLink`
- `SetGroupLinkMemberRole`
- `DeleteGroupLink`
- `GetGroupLink`
- `EnableAddressAutoAccept`
- `SendTextMessage`
- `SendTextMessageWithOptions`
- `SendTextToContact`
- `SendTextToContactWithOptions`
- `SendTextToGroup`
- `SendTextToGroupWithOptions`
- `UpdateChatItem`
- `UpdateTextMessage`
- `UpdateTextMessageInContact`
- `UpdateTextMessageInGroup`
- `DeleteChatItems`
- `DeleteChatItemsInContact`
- `DeleteChatItemsInGroup`
- `ModerateDeleteGroupChatItems`
- `DeleteChat`
- `DeleteContactChat`
- `DeleteGroupChat`
- `SetChatItemReaction`
- `AddChatItemReaction`
- `RemoveChatItemReaction`
- `ReceiveFile`
- `CancelFile`

Runnable example:

```bash
go run ./examples/echo
```

The generator currently reads:

- `spec/upstream/COMMANDS.md`
- `spec/upstream/commands.ts`
- `spec/upstream/events.ts`
- `spec/upstream/responses.ts`

and produces:

- `sdk/command/generated_catalog.go`
- `sdk/command/generated_requests.go`
- `sdk/client/generated_senders.go`
- `sdk/types/generated_tags.go`
- `sdk/types/generated_records.go`

To refresh snapshots from upstream and regenerate in one step:

```bash
./scripts/update-upstream.sh
```

Optional branch/ref:

```bash
./scripts/update-upstream.sh stable
```
