# Proof server of NextID

## What's this?

A bridge to connect your web2.0 service / cryptocurrency wallet to
NextID identity system. So your friends or followers can easily find
you on any supported platform.

## Which platform can I prove?

### Supported

| Platform    | `platform` value | `identity` value             | `proof_location` value                                                         | Misc.                                                   |
|-------------|------------------|------------------------------|--------------------------------------------------------------------------------|---------------------------------------------------------|
| Twitter     | `twitter`        | `twitter_username`           | Proof tweet ID (`1415362679095635970`)                                         |                                                         |
| Keybase     | `keybase`        | `keybase_username`           | N/A (use `https://your_identity.keybase.pub/NextID/COMPRESSED_PUBKEY_HEX.txt`) |                                                         |
| Ethereum    | `ethereum`       | Wallet address `0x123AbC...` | N/A (Two-way signatures created from persona sk and wallet sk)                 |                                                         |
| Github      | `github`         | `github_username`            | Public visible Gist ID `a6dddd2811af21b671fd`                                  | Gist should contain `0xPUBKEY_COMRESSED_HEX.json` file |
| Discord     | `discord`        | `UserName#0000`              | message link (`https://discord.com/channels/DIGITS/DIGITS/DIGITS`)             |                                                         |
| DotBit      | `dotbit`         | `address.bit`                | Custom type Record (`nextid_proof_0xPUBKEY_COMRESSED_HEX`)                     | Formerly known as DAS (Decentralized Account System)    |
| Solana      | `solana`         | Wallet address `AbCdEfG9...` | N/A (Two-way signatures created from persona sk and wallet sk)                 |                                                         |
| Minds       | `minds`          | `minds_username`             | Proof post ID (`LONG_DIGITS` in `https://www.minds.com/newsfeed/LONG_DIGITS`)  |                                                         |
| DNS         | `dns`            | `example.com`                | N/A (use `dig example.com TXT`)                                                |                                                         |
| ActivityPub | `activitypub`    | `username@server.com`        | ID-ish string in "toot"'s detail page link                                     | Supports `mastodon`, `pleroma` and `misskey` instances  |

### Planning

| Platform                             | `platform` value | `identity` value                                                  | `proof_location` value                                 | Misc. |
|--------------------------------------|------------------|-------------------------------------------------------------------|--------------------------------------------------------|-------|
| [Facebook](https://www.facebook.com) | `facebook`       | Username in link (i.e. `Meta` in `https://www.facebook.com/Meta`) | Post ID (`460695145492083`)                            |       |
| [Telegram](https://telegram.org)     | `telegram`       | `telegram`                                                        | `https://t.me/some_public_group/CHAT_ID_DIGITS`        |       |
| ENS                                  | `ens`            | `myens.eth`                                                       | N/A (use `id.next.proof` record in ENS to store proof) |       |
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

### Can this be decentralized?

TLDR: If it is, the whole system will become weak against "junk proof"
attack. For example:

Can an attacker claims himself as `@elonmusk` (and provides a
seem-to-be-normal proof tweet link)?

Of course he can, since this network is decentralized, there is no
"gatekeeper", anybody can publish any data in their own namespace.

Well, how should other users trust this claim?  They can only fetch this
tweet and validate it locally.

The result is:

1. EVERY user should set fetch methods for EVERY platform available
   (typically API key) for their own.  Since they can only trust their
   own verification result.  This will increase the difficulty of
   deployment for every single user in this network.

2. With "junk proofs" growing more and more (they will not disappear or
   be deleted by someone, decentralized, you know), every user
   will waste more and more API usage on these unsuccessful proofs.
   In the end the whole network is flooded.

## Presentation

- <2022-01-11 Tue> :: (Chinese) [A brief introduction of ProofService & NextID](https://docs.google.com/presentation/d/1aq9H8eViLRgZ32xcTcTsAdET52X3P3jtuJFIP5COpyI/edit?usp=sharing)
