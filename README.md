# Crypto-SMS

Crypto-SMS is a backend service that enables secure cryptocurrency transactions via SMS without the need for internet connectivity. Users can send SMS messages to a Twilio phone number, and the backend processes the transactions, sends confirmation messages, and manages user data.

## Features

- **Send and Receive SMS Transactions**: Users can send cryptocurrency transactions via SMS, and receive confirmation messages.
- **2-Factor Authentication**: Generate and verify 2FA codes for enhanced security.
- **Manage Phone Numbers and Wallets**: Update phone numbers linked to wallet addresses, and check existing linkages.
- **View Balances**: Fetch and view all cryptocurrency balances for a given wallet address.

## Endpoints

### Twilio Webhook

Handles incoming SMS messages from Twilio.

```http
POST /twilio-webhook
