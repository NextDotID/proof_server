# Proof server of NextID

## What's this?

A bridge to connect your web2.0 service / cryptocurrency wallet to
NextID identity system. So your friends or followers can easily find
you on any supported platform.

## Which platform can I prove?

### Supported

| Platform | `platform` value | `identity` value             | `proof_location` value                                                         | Misc.                                                   |
|----------|------------------|------------------------------|--------------------------------------------------------------------------------|---------------------------------------------------------|
| Twitter  | `twitter`        | `twitter_username`           | Proof tweet ID (`1415362679095635970`)                                         |                                                         |
| Keybase  | `keybase`        | `keybase_username`           | N/A (use `https://your_identity.keybase.pub/NextID/COMPRESSED_PUBKEY_HEX.txt`) |                                                         |
| Ethereum | `eth`            | Wallet address `0x123AbC...` | N/A (Two-way signatures created from persona sk and wallet sk)                 |                                                         |
| Github   | `github`         | `github_username`            | Public visible Gist ID `a6dddd2811af21b671fd`                                  | Gist should contains `0xPUBKEY_COMRESSED_HEX.json` file |

### Planning

| Platform                             | `platform` value | `identity` value                                                  | `proof_location` value                                 | Misc. |
|--------------------------------------|------------------|-------------------------------------------------------------------|--------------------------------------------------------|-------|
| [Facebook](https://www.facebook.com) | `facebook`       | Username in link (i.e. `Meta` in `https://www.facebook.com/Meta`) | Post ID (`460695145492083`)                            |       |
| [Minds](https://www.minds.com)       | `minds`          | `minds`                                                           | Newsfeed ID (`1309718521097228301`)                    |       |
| [Telegram](https://telegram.org)     | `telegram`       | `telegram`                                                        | `https://t.me/some_public_group/CHAT_ID_DIGITS`        |       |
| [Discord](https://discord.com)       | `discord`        | `Discord#0000`                                                    | `https://discord.com/channels/DIGITS/DIGITS/DIGITS`    |       |
| DNS (TXT field)                      | `dns`            | `example.com`                                                     | N/A (use `dig example.com TXT`)                        |       |
| ENS                                  | `ens`            | `myens.eth`                                                       | N/A (use `id.next.proof` record in ENS to store proof) |       |
| Decentrialized Account Systems       | `das`            |                                                                   |                                                        |       |
| Email                                | `email`          | `mail_address@example.com`                                        | A public mailing list `mbox` download URL (?)          |       |

## How?

### Proof Chain

Each NextID identity (named **persona**) has its own "proof chain":

1. Every modification of status (add / delete a proof) will become a
   "link" in the chain.

2. Every link is signed by persona owner (aka persona private key).

   > Furthermore, each link signature payload will contain previous
   > link's sig, so the whole chain cannot be modified other than
   > owner. Kinda like "blockchain".

3. Every persona proof chain is public. Everyone can download and
   verify it manually.

So you don't need to trust our hosted server, since we cannot fake a
proof modification for any persona.

> See [HERE](./proof_chain.md) for structure of this proof chain file.

## FAQ

### Can this be decentrialized?

TLDR: If it is, the whole system will become weak against "junk proof"
attack. For example:

Can an attacker claims himself as `@elonmusk` (and provides an
seem-to-be-normal proof tweet link)?

Of course he can, since this network is decentrialized, there is no
"gatekeeper", anybody can publish any data in their own namespace.

Well, how should other users trust this claim?  They can only fetch this
tweet and validate it locally.

The result is:

1. EVERY user should set fetch methods for EVERY platform available
   (typically API key) for their own.  Since they can only trust their
   own verfification result.  This will increase the difficulty of
   deployment for every single user in this network.

2. With "junk proofs" grow more and more (they will not disappear or
   being deleted by someone, decentrialized, you know), every user
   will waste more and more API usage on these unsuccessful proof.
   In the end the whole network is flooded.
