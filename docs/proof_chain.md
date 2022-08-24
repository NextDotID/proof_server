# Structure of Proof Chain

## Type declaration

```typescript
const VERSION = "1";

// assert(signature.match(/0x[a-f0-9]{144}/))
// Sample:
// 0x3046022100881328457aa312135c37e1ddf8a129717274ce3f389c176936f5cb44edf04fc4022100be183139154d108ce2e5d6ba16678b0dbeb3b7d70caac2b00b2dad8f81e87790
type Signature = string;

// assert(public_key.match(/^0x[a-f0-9]{130}$/))
// Sample:
// 0x0428b73a2b67a88a47edb15bed5c73a199e24287bb12997c54239e9e6815e24a3032a502d58afe3f36a54f2f7606022907f358d0dd58939cffa0a845c5043ce038
type PublicKey = string;

// All available chain modification actions
enum Action {
    Create = "create",
    Delete = "delete",
}

// All supported platforms
enum Platform {
    Twitter = "twitter",
    Keybase = "keybase",
}

// Every link in the proof chain
interface Link {
    // If this is genesis link, leave it null; else, it equals
    // previous link's signature. Worked as a pointer.
    prev: Signature | null;
    action: Action;
    platform: Platform;
    identity: string;
    // if method === Method.Add, then it must be a string; else, left null
    proof_location: string | null;
    // UNIX timestamp (unit: second)
    created_at: number;
    // Signature of self
    signature: Signature;
}

// Main struct
interface Chain {
    version: VERSION;
    avatar: {
        public_key: PublicKey,
        curve: "secp256k1",
    };
    links: Link[];
}
```


## Example

```javascript
{
    "version": "1",
    "avatar": {
        "public_key": "0x0485554db28de6fefb7fe532164b67372a5e9d78dfd7f77e09a8b274f777c3e64f2e20353df005a83dbe4c5ca663638621ce4d1dd0c9586ab3fc71286b74741ed8",
        "curve": "secp256k1"
    },
    "links": [{
        "prev": null, // Genesis
        "action": "create",
        "platform": "twitter",
        "identity": "twitter_screen_name",
        "proof_location": "https://twitter.com/twitter_screen_name/11111111111111",
        "created_at": 1638618231,
        "signature": "0xSIG1"
    }, {
        "prev": "0xSIG1",
        "action": "create",
        "platform": "keybase",
        "identity": "keybase_username",
        "proof_location": "https://keybase_username.keybase.pub/NextID/proof.txt",
        "created_at": 1638618470,
        "signature": "0xSIG2"
    }]
}
```

## How to sign

```typescript
// Pseudo-code for how to sign a link
function sign_link(link: Link): Signature {
    // Omit "signature" and "proof_location" KV from original link
    let signature_payload_struct = Object.assign({}, link);
    delete(signature_payload_struct.signature);
    delete(signature_payload_struct.proof_location);

    // Sort this object by key
    const signature_payload_struct_sorted = sort_key(signature_payload_struct);

    // JSONify
    const signature_payload = JSON.stringify(signature_payload_struct_sorted);

    // Sign this using web3 personal_sign method
    // Specifically:
    let personal_signature_payload = keccak256("\x19Ethereum Signed Message:\n" + signature_payload.length + signature_payload)
    let signature_bin: Buffer = avatar_private_key.sign(personal_signature_payload)
    let signature = "0x" + Base16.encode(signature_bin, {case: 'lower'})

    // const signature = web3.eth.personal.sign(signature_payload, avatar_private_key);

    // Final artifact should be a format like below:
    assert(signature.match(/^0x[0-9a-f]{130}$/))
    return signature;
}
```
