# Notification subsystem

Original motivation, requirements and general solution strategy are described
in the issue #895.

This extension allows a websocket client to subscribe to various events and
receive them as JSON-RPC notifications from the server.

## Events
Currently supported events:
 * new block added
   Contents: block.
   Filters: primary ID.
 * new transaction in the block
   Contents: transaction.
   Filters: sender and cosigner.
 * notification generated during execution
   Contents: container hash, contract script hash, stack item.
   Filters: contract script hash.
 * transaction executed
   Contents: application execution result.
   Filters: VM state.

Filters use conjunctional logic.

## Ordering and persistence guarantees
 * new block is only announced after its processing is complete and the chain
   is updated to the new height
 * no disk-level persistence guarantees are given
 * new in-block transaction is announced after block processing, but before
   announcing the block itself
 * transaction notifications are only announced for successful transactions
 * all announcements are being done in the same order they happen on the chain
   At first transaction execution is announced, then followed by notifications
   generated during this execution, then followed by transaction announcement.
   Transaction announcements are ordered the same way they're in the block.
 * unsubscription may not cancel pending, but not yet sent events

## Subscription management

To receive events clients need to subscribe to them first via `subscribe`
method. Upon successful subscription clients receive subscription ID for
subsequent management of this subscription. Subscription is only valid for
connection lifetime, no long-term client identification is being made.

Errors are not described down below, but they can be returned as standard
JSON-RPC errors (most often caused by invalid parameters).

### `subscribe` method

Parameters: event stream name, stream-specific filter rules hash (can be
omitted if empty).

Recognized stream names:
 * `block_added`
   Filter: `primary` as an integer with primary (speaker) node index from
   ConsensusData.
 * `transaction_added`
   Filter: `sender` field containing string with hex-encoded Uint160 (LE
   representation) for transaction's `Sender` and/or `cosigner` in the same
   format for one of transaction's `Cosigners`.
 * `notification_from_execution`
   Filter: `contract` field containing string with hex-encoded Uint160 (LE
   representation).
 * `transaction_executed`
   Filter: `state` field containing `HALT` or `FAULT` string for successful
   and failed executions respectively.

Response: returns subscription ID (string) as a result. This ID can be used to
cancel this subscription and has no meaning other than that.

Example request (subscribe to notifications from contract
0x6293a440ed80a427038e175a507d3def1e04fb67 generated when executing
transactions):

```
{
  "jsonrpc": "2.0",
  "method": "subscribe",
  "params": ["notification_from_execution", {"contract": "6293a440ed80a427038e175a507d3def1e04fb67"}],
  "id": 1
}

```

Example response:

```
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": "55aaff00"
}
```

### `unsubscribe` method

Parameters: subscription ID as a string.

Response: boolean true.

Example request (unsubscribe from "55aaff00"):

```
{
  "jsonrpc": "2.0",
  "method": "unsubscribe",
  "params": ["55aaff00"],
  "id": 1
}
```

Example response:

```
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": true
}
```

## Events

Events are sent as JSON-RPC notifications from the server with `method` field
being used for notification names. Notification names are identical to stream
names described for `subscribe` method with one important addition for
`event_missed` which can be sent for any subscription to signify that some
events were not delivered (usually when client isn't able to keep up with
event flow).

Verbose responses for various structures like blocks and transactions are used
to simplify working with notifications on client side. Returned structures
mostly follow the one used by standard Neo RPC calls, but may have some minor
differences.

If a server-side event matches several subscriptions from one client, it's
only sent once.

### `block_added` notification
As a first parameter (`params` section) contains block converted to JSON
structure which is similar to verbose `getblock` response but with the
following differences:
 * it doesn't have `size` field (you can calculate it client-side)
 * it doesn't have `nextblockhash` field (it's supposed to be the latest one
    anyway)
 * it doesn't have `confirmations` field (see previous)

No other parameters are sent.

Example:
```
{
   "params" : [
      {
         "index" : 207,
         "time" : 1590006200,
         "nextconsensus" : "AXSvJVzydxXuL9da4GVwK25zdesCrVKkHL",
         "consensus_data" : {
            "primary" : 0,
            "nonce" : "0000000000000457"
         },
         "previousblockhash" : "0x04f7580b111ec75f0ce68d3a9fd70a0544b4521b4a98541694d8575c548b759e",
         "witnesses" : [
            {
               "invocation" : "0c4063429fca5ff75c964d9e38179c75978e33f8174d91a780c2e825265cf2447281594afdd5f3e216dcaf5ff0693aec83f415996cf224454495495f6bd0a4c5d08f0c4099680903a954278580d8533121c2cd3e53a089817b6a784901ec06178a60b5f1da6e70422bdcadc89029767e08d66ce4180b99334cb2d42f42e4216394af15920c4067d5e362189e48839a24e187c59d46f5d9db862c8a029777f1548b19632bfdc73ad373827ed02369f925e89c2303b64e6b9838dca229949b9b9d3bd4c0c3ed8f0c4021d4c00d4522805883f1db929554441bcbbee127c48f6b7feeeb69a72a78c7f0a75011663e239c0820ef903f36168f42936de10f0ef20681cb735a4b53d0390f",
               "verification" : "130c2102103a7f7dd016558597f7960d27c516a4394fd968b9e65155eb4b013e4040406e0c2102a7bc55fe8684e0119768d104ba30795bdcc86619e864add26156723ed185cd620c2102b3622bf4017bdfe317c58aed5f4c753f206b7db896046fa7d774bbc4bf7f8dc20c2103d90c07df63e690ce77912e10ab51acc944b66860237b608c4f8f8309e71ee699140b413073b3bb"
            }
         ],
         "version" : 0,
         "hash" : "0x239fea00c54c2f6812612874183b72bef4473fcdf68bf8da08d74fd5b6cab030",
         "tx" : [
            {
               "txid" : "0xf736cd91ab84062a21a09b424346b241987f6245ffe8c2b2db39d595c3c222f7",
               "scripts" : [
                  {
                     "verification" : "0c2102b3622bf4017bdfe317c58aed5f4c753f206b7db896046fa7d774bbc4bf7f8dc20b410a906ad4",
                     "invocation" : "0c4016e7a112742409cdfaad89dcdbcb52c94c5c1a69dfe5d8b999649eaaa787e31ca496d1734d6ea606c749ad36e9a88892240ae59e0efa7f544e0692124898d512"
                  }
               ],
               "vout" : [],
               "cosigners" : [],
               "valid_until_block" : 1200,
               "nonce" : 8,
               "net_fee" : "0.0030421",
               "sender" : "ALHF9wsXZVEuCGgmDA6ZNsCLtrb4A1g4yG",
               "sys_fee" : "0",
               "type" : "InvocationTransaction",
               "attributes" : [],
               "version" : 1,
               "vin" : [],
               "size" : 204,
               "script" : "10c00c04696e69740c14769162241eedf97c2481652adf1ba0f5bf57431b41627d5b52"
            },
            {
               "script" : "01e8030c14316e851039019d39dfc2c37d6c3fee19fd5809870c14769162241eedf97c2481652adf1ba0f5bf57431b13c00c087472616e736665720c14769162241eedf97c2481652adf1ba0f5bf57431b41627d5b5238",
               "size" : 277,
               "attributes" : [],
               "version" : 1,
               "vin" : [],
               "net_fee" : "0.0037721",
               "sender" : "ALHF9wsXZVEuCGgmDA6ZNsCLtrb4A1g4yG",
               "sys_fee" : "0",
               "type" : "InvocationTransaction",
               "nonce" : 9,
               "cosigners" : [
                  {
                     "scopes" : 1,
                     "account" : "0x870958fd19ee3f6c7dc3c2df399d013910856e31"
                  }
               ],
               "valid_until_block" : 1200,
               "scripts" : [
                  {
                     "invocation" : "0c4027727296b84853c5d9e07fb8a40e885246ae25641383b16eefbe92027ecb1635b794aacf6bbfc3e828c73829b14791c483d19eb758b57638e3191393dbf2d288",
                     "verification" : "0c2102b3622bf4017bdfe317c58aed5f4c753f206b7db896046fa7d774bbc4bf7f8dc20b410a906ad4"
                  }
               ],
               "vout" : [],
               "txid" : "0xe1cd5e57e721d2a2e05fb1f08721b12057b25ab1dd7fd0f33ee1639932fdfad7"
            }
         ],
         "merkleroot" : "0xb2c7230ebee4cb83bc03afadbba413e6bca8fcdeaf9c077bea060918da0e52a1"
      }
   ],
   "jsonrpc" : "2.0",
   "method" : "block_added"
}
```

### `transaction_added` notification

In the first parameter (`params` section) contains transaction converted to
JSON which is similar to verbose `getrawtransaction` response, but with the
following differences:
 * block's metadata is missing (`blockhash`, `confirmations`, `blocktime`)

No other parameters are sent.

Example:
```
{
   "method" : "transaction_added",
   "params" : [
      {
         "valid_until_block" : 1200,
         "version" : 1,
         "txid" : "0xe1cd5e57e721d2a2e05fb1f08721b12057b25ab1dd7fd0f33ee1639932fdfad7",
         "scripts" : [
            {
               "invocation" : "0c4027727296b84853c5d9e07fb8a40e885246ae25641383b16eefbe92027ecb1635b794aacf6bbfc3e828c73829b14791c483d19eb758b57638e3191393dbf2d288",
               "verification" : "0c2102b3622bf4017bdfe317c58aed5f4c753f206b7db896046fa7d774bbc4bf7f8dc20b410a906ad4"
            }
         ],
         "sys_fee" : "0",
         "sender" : "ALHF9wsXZVEuCGgmDA6ZNsCLtrb4A1g4yG",
         "vout" : [],
         "net_fee" : "0.0037721",
         "size" : 277,
         "attributes" : [],
         "script" : "01e8030c14316e851039019d39dfc2c37d6c3fee19fd5809870c14769162241eedf97c2481652adf1ba0f5bf57431b13c00c087472616e736665720c14769162241eedf97c2481652adf1ba0f5bf57431b41627d5b5238",
         "nonce" : 9,
         "vin" : [],
         "type" : "InvocationTransaction",
         "cosigners" : [
            {
               "account" : "0x870958fd19ee3f6c7dc3c2df399d013910856e31",
               "scopes" : 1
            }
         ]
      }
   ],
   "jsonrpc" : "2.0"
}
```

### `notification_from_execution` notification

Contains three parameters: container hash (hex-encoded LE Uint256 in a
string), contract script hash (hex-encoded LE Uint160 in a string) and stack
item (encoded the same way as `state` field contents for notifications from
`getapplicationlog` response).

Example:

```
{
   "jsonrpc" : "2.0",
   "method" : "notification_from_execution",
   "params" : [
      {
         "state" : {
            "value" : [
               {
                  "value" : "636f6e74726163742063616c6c",
                  "type" : "ByteArray"
               },
               {
                  "value" : "7472616e73666572",
                  "type" : "ByteArray"
               },
               {
                  "value" : [
                     {
                        "value" : "769162241eedf97c2481652adf1ba0f5bf57431b",
                        "type" : "ByteArray"
                     },
                     {
                        "value" : "316e851039019d39dfc2c37d6c3fee19fd580987",
                        "type" : "ByteArray"
                     },
                     {
                        "value" : "1000",
                        "type" : "Integer"
                     }
                  ],
                  "type" : "Array"
               }
            ],
            "type" : "Array"
         },
         "contract" : "0x1b4357bff5a01bdf2a6581247cf9ed1e24629176"
      }
   ]
}
```

### `transaction_executed` notification

Contains the same result as from `getapplicationlog` method in the first
parameter and no other parameters. One difference from `getapplicationlog` is
that it always contains zero in the `contract` field.

Example:
```
{
   "method" : "transaction_executed",
   "params" : [
      {
         "txid" : "0xe1cd5e57e721d2a2e05fb1f08721b12057b25ab1dd7fd0f33ee1639932fdfad7",
         "executions" : [
            {
               "trigger" : "Application",
               "gas_consumed" : "2.291",
               "contract" : "0x0000000000000000000000000000000000000000",
               "stack" : [],
               "notifications" : [
                  {
                     "state" : {
                        "type" : "Array",
                        "value" : [
                           {
                              "value" : "636f6e74726163742063616c6c",
                              "type" : "ByteArray"
                           },
                           {
                              "type" : "ByteArray",
                              "value" : "7472616e73666572"
                           },
                           {
                              "value" : [
                                 {
                                    "value" : "769162241eedf97c2481652adf1ba0f5bf57431b",
                                    "type" : "ByteArray"
                                 },
                                 {
                                    "type" : "ByteArray",
                                    "value" : "316e851039019d39dfc2c37d6c3fee19fd580987"
                                 },
                                 {
                                    "value" : "1000",
                                    "type" : "Integer"
                                 }
                              ],
                              "type" : "Array"
                           }
                        ]
                     },
                     "contract" : "0x1b4357bff5a01bdf2a6581247cf9ed1e24629176"
                  },
                  {
                     "contract" : "0x1b4357bff5a01bdf2a6581247cf9ed1e24629176",
                     "state" : {
                        "value" : [
                           {
                              "value" : "7472616e73666572",
                              "type" : "ByteArray"
                           },
                           {
                              "value" : "769162241eedf97c2481652adf1ba0f5bf57431b",
                              "type" : "ByteArray"
                           },
                           {
                              "value" : "316e851039019d39dfc2c37d6c3fee19fd580987",
                              "type" : "ByteArray"
                           },
                           {
                              "value" : "1000",
                              "type" : "Integer"
                           }
                        ],
                        "type" : "Array"
                     }
                  }
               ],
               "vmstate" : "HALT"
            }
         ]
      }
   ],
   "jsonrpc" : "2.0"
}
```

### `event_missed` notification

Never has any parameters. Example:

```
{
  "jsonrpc": "2.0",
  "method": "event_missed",
  "params": []
}
```
