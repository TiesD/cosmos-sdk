# ADR 074: Messages with Implicit Signers

## Changelog

* 2024-06-10: Initial draft

## Status

PROPOSED Not Implemented

## Abstract

> "If you can't explain it simply, you don't understand it well enough." Provide
> a simplified and layman-accessible explanation of the ADR.
> A short (~200 word) description of the issue being addressed.

## Context

> This section describes the forces at play, including technological, political,
> social, and project local. These forces are probably in tension, and should be
> called out as such. The language in this section is value-neutral. It is simply
> describing facts. It should clearly explain the problem and motivation that the
> proposal aims to resolve.
> {context body}

Historically operations in the SDK have been modelled with the `sdk.Msg` interface and
the account signing the message has to be explicitly extracted from the body of `Msg`s.
Originally this was via a `GetSigners` method on the `sdk.Msg` interface which returned
instances of `sdk.AccAddress` which itself relied on a global variable for decoding
the addresses from bech32 strings. This was a messy situation. In addition, the implementation
for `GetSigners` was different for each `Msg` type and clients would need to do a custom
implementation for each `Msg` type. These were improved somewhat with the introduction of
the `cosmos.msg.v1.signer` protobuf option which allowed for a more standardised way of
defining who the signer of a message was and its implementation in the `x/tx` module which
extracts signers dynamically and allowed removing the dependency on the global bech32
configuration.

Still this design introduces a fair amount of complexity. For instance, inter-module message
passing ([ADR 033](./adr-033-protobuf-inter-module-comm.md)) has been in discussion for years
without much progress and one of the main blockers is figuring out how to properly authenticate
messages in a performant and consistent way. With embedded message signers there will always need
to be a step of extracting the signer and then checking with the module sending is actually
authorized to perform the operation. With dynamic signer extraction, although the system is
more consistent, more performance overhead is introduced. In any case why should an inter-module
message passing system need to do so much conversion, parsing, etc. just to check if a message
is authenticated? In addition, we have the complexity where modules can actually have many valid
addresses. How are we to accommodate this? Should there be a lookup into `x/auth` to check if an
address belongs to a module or not? All of these thorny questions are delaying the delivery of
inter-module message passing because we do not want an implementation that is overly complex.
There are many use cases for inter-module message passing which are still relevant, the most
immediate of which is a more robust denom management system in `x/bank` `v2` which is being explored
in [ADR 071](https://github.com/cosmos/cosmos-sdk/pull/20316).

## Alternatives

Alternatives that have been considered are extending the current `x/tx` signer extraction system
to inter-module message passing as defined in [ADR 033](./adr-033-protobuf-inter-module-comm.md).

## Decision

We have decided to introduce a new `MsgV2` standard whereby the signer of the message is implied
by the credentials of the party sending it. These messages will be distinct from the existing messages
and define new semantics with the understanding that signers are implicit.

In the case of messages passed internally by a module or `x/account` instance, the signer of a message
will simply be the main root address of the module or account sending the message. An interface for
safely passing such messages to the message router will need to be defined.

In the case of messages passed externally in transactions, `MsgV2` instances will need to be wrapped
in a `MsgV2` envelope:
```protobuf
message MsgV2 {
  string signer = 1;
  google.protobuf.Any msg = 2;  
}
```

Because the `cosmos.msg.v1.signer` annotation is required currently, `MsgV2` types should set the message option
`cosmos.msg.v2.is_msg` to `true` instead.

Modules defining handlers for `MsgV2` instances will need to extract the sender from the `context.Context` that is
passed in. An interface in `core` which will be present on the `appmodule.Environment` will be defined for this purpose:
```go
type GetSenderService interface {
  GetSender(ctx sdk.Context) []byte
}
```

Sender addresses that are returned by the service will be simple `[]byte` slices and any bech32 conversion will be
done by the framework.

## Consequences

### Backwards Compatibility

This design does not depreciate the existing method of embedded signers in `Msg`s and is totally compatible with it.

### Positive

* Allows for a simple inter-module communication design which can be used soon for the `bank` `v2` redesign.
* Allows for simpler client implementations for messages in the future.

### Negative

* There will be two message designs and developers will need to pick between them.

### Neutral

## Further Discussions

Further discussions should take place on GitHub.

## References

* [ADR 033](./adr-033-protobuf-inter-module-comm.md)