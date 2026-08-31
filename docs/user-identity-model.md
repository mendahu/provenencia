# Provenencia Genealogy — User and Contributor Identity Model

## Status

Draft architecture notes. This document defines the MVP contributor identity model and the extension boundaries for future synchronization, sharing, hosted backup, and web access.

The MVP requirement is deliberately small:

> A researcher must be able to create and use a Provenencia project locally without authentication, while every audited research change has a durable contributor identity that can later participate in a broader collaboration model without rewriting history.

---

# 1. MVP scope

The MVP supports a single local user on a desktop installation.

There is no required login, cloud account, project membership system, role system, public-key identity, or external identity provider.

The only current identity requirement is durable audit attribution.

```text
Who made this research change?
```

That question is answered by a globally unique User UUID and a display name.

---

# 2. Local-first identity

Contributor identity is a UUIDv7, a display name, and a short `USR-…` ref. It is created on **first create-project or first open** on an install (mint a new User, or adopt one already in the folder). It is stored in **application support** and copied into each project’s `users` table.

For example:

```text
User U1
  id = 019c...
  display_name = "Jake Robins"
  ref = "USR-F4N2P"
```

That UUID becomes the durable contributor identity used by audit history.

Creating a project records the local User in the project, and ordinary audited writes reference that User.

```text
Revision 1
  user_id = U1
  action = create_project
```

No email address, password, network connection, or online service is required.

The desktop application should reuse the same local User UUID when the researcher creates additional projects. This keeps the identity globally stable rather than generating a new contributor identity per project.

The **project file does not record which machine it is on.** The OS account is not an identity provider. The install stores “I am `U1`” in **application support** (outside the project directory). Opening a project compares that to `users` / audit history in the folder.

```text
First create-project or first open on this install
  → mint or adopt a User UUID
  → persist it in application support for later compares

Open, local UUID matches a User already in the project
  → keep writing as that User (e.g. new MacBook after copying *your* folder)

Open, no local UUID yet, project already has User U1
  → ask: “Continue as Jake Robins?” → adopt U1 into application support
  → or “That’s not me” → mint U2, insert into this project’s `users`, new writes as U2
     (U1’s history stays; both live in the same file)
```

A spouse opening a copied folder on her machine chooses “that’s not me,” gets her own UUID on **her** install, and her revisions sit beside the original contributor’s. Cloud login later **maps onto** an existing `users.id`; it does not replace it ([`application-stack.md`](application-stack.md) §17).

---

# 3. Current `users` schema

The MVP `users` table is intentionally minimal.

```sql
CREATE TABLE users (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    display_name    TEXT NOT NULL,
    ref             TEXT UNIQUE                -- USR-F4N2P; required by the app
) STRICT;
```

## 3.1 `id`

`id` is the durable contributor identity.

It is a globally unique UUIDv7 rather than a project-local integer, email address, username, or machine identifier.

That choice is the primary extensibility mechanism. Future synchronization can transport the same User UUID between replicas without changing historical `audit_transactions.user_id` references.

## 3.2 `display_name`

`display_name` is the human-readable attribution shown in research history and other UI.

It need not be unique and is not authentication data.

Examples:

```text
Jake Robins
Jane Smith
A. García
```

If the display name changes, that durable project-data change is itself auditable.

## 3.3 `ref`

`ref` is a short user-facing identifier in the shared `{PREFIX}-{token}` form with prefix `USR` (for example `USR-F4N2P`). It distinguishes contributors who share a display name. Machine identity remains the UUID. The same `ref` is stored on the install-local `identity.json` and copied into each project’s `users` row.

---

# 4. Relationship to audit history

`audit_transactions.user_id` references `users.id`.

```sql
audit_transactions.user_id BLOB REFERENCES users(id)
```

The User row is the authoritative attribution identity for research changes.

For example:

```text
Revision 142
User: Jake Robins
Action: replace_artifact_file
```

Because the display name is stored in the project, the audit history remains understandable offline.

Users referenced by audit history should not be destructively deleted in ordinary application behavior because historical attribution must remain resolvable.

---

# 5. Identity is not authentication

Even though the MVP has no authentication system, the design should preserve a clear boundary between contributor identity and login/authentication concepts.

The portable project database should not acquire authentication secrets later simply because collaboration is added.

Do not store project-level fields such as:

```text
password_hash
access_token
refresh_token
OAuth client secret
session cookie
passkey private material
```

A future desktop authentication or synchronization layer may use operating-system secure storage. A hosted service may maintain its own authentication database.

The project User UUID remains the research attribution identity regardless of how a future service authenticates the person using it.

---

# 6. Future extension boundaries

The following concepts are intentionally **not part of the MVP schema**, but the current User UUID model is designed so they can be added later without changing existing audit history.

## 6.1 External identity mappings

A future hosted or third-party identity may be linked to an existing User.

A possible future table is:

```sql
CREATE TABLE user_external_identities (
    id                  BLOB PRIMARY KEY,
    user_id             BLOB NOT NULL REFERENCES users(id),
    provider            TEXT NOT NULL,
    provider_subject    TEXT NOT NULL,
    display_identifier  TEXT,

    UNIQUE (provider, provider_subject)
) STRICT;
```

This is **not an MVP table**.

The important principle is that an external account would extend an existing `users.id`; it would not replace the User UUID already referenced by audit history.

## 6.2 Project membership and authorization

Future collaboration may require concepts such as:

```text
owner
editor
viewer
```

These belong to a project-membership or synchronization authorization model, not to `users`.

A User answers:

```text
Who is this contributor?
```

A future membership model answers:

```text
What may this contributor do in this project?
```

The MVP does not need a membership/roles table. A copied project may already contain more than one User; that is attribution, not authorization.

## 6.3 Cryptographic/synchronization identity

A future peer-to-peer or untrusted sync system may require contributors to prove ownership of their User identity using a public/private keypair.

That should be layered onto the existing User UUID rather than embedded in the MVP identity record before there is a concrete synchronization design.

Private key material must not be stored in the portable project database.

## 6.4 Identity reconciliation

Future synchronization may encounter two independently created User UUIDs that represent the same human.

Those identities must not be silently merged based on display names or email addresses. A future explicit reconciliation/alias mechanism can handle that while preserving original audit references.

---

# 7. Email and profile information

The MVP does not store email addresses or conventional account-profile fields.

Do not add fields such as:

```text
email
avatar
organization
biography
website
```

without a concrete product requirement.

In particular, email should not be used as the User identity because it is mutable, unnecessary for local use, and may differ between future services.

---

# 8. MVP lifecycle

```text
1. Researcher installs Provenencia.
2. First create-project or first open: mint or adopt a User UUID; store it in application support; ask for a display name if minting.
3. Create-project copies that User into the project; open-project compares install UUID to users in the folder (match / “that’s me” / “not me”).
4. Audit transactions reference the writer’s User UUID.
5. No authentication or collaboration infrastructure is involved.
```

If collaboration is added later:

```text
8. Existing User UUID remains unchanged.
9. New identity/authentication/membership mechanisms attach to it.
10. Existing audit history requires no migration of contributor identity.
```

---

# 9. Current schema

```text
users
  id
  display_name

users
  ↑
  │ audit attribution
  │
audit_transactions
```

Everything else in this document's future-extension section is intentionally deferred.

---

# 10. Current architectural rules

1. MVP supports one local user per desktop application identity.
2. No authentication or network access is required.
3. `users` contains `id`, `display_name`, and `ref` (`USR-…`) in the MVP.
4. User IDs are globally unique UUIDv7 values rather than project-local integers.
5. The same local User UUID should be reused across projects created by that desktop identity.
6. `audit_transactions.user_id` references the durable User UUID.
7. Display names are descriptive and need not be unique.
8. Users referenced by audit history are preserved.
9. Email is not required and is not an identity key.
10. Authentication credentials do not belong in portable project databases.
11. External identities, membership/roles, cryptographic identity, and identity reconciliation are future concerns rather than MVP schema.
12. Future identity mechanisms should extend the existing User UUID rather than replacing it.
13. All tables use SQLite `STRICT` typing.

This keeps the MVP implementation small while preserving the one decision that is difficult to retrofit later: a stable, globally unique contributor identity for audit history.