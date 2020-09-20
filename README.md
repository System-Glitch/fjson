# FJSON - Fast JSON

A prototype for a fast network protocol based on JSON exchanges.

## Protocol

All exchanges are expected to be:
- Strings
- Null-terminated
- Encoded in UTF-8
- Valid JSON

Connections are not long-lived, not kept alive. Once the server has received a valid request and sent a response, the connection is closed. In case of error such as invalid payload, the connection is simply closed.