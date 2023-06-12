# go-wingman

Backend service for Wingman dating service.

Core features include:
  - Session based authentication
  - Integration with APNS for sending notification to ios clients
  - Caching system for sessions and certain non-volatile data

Structure:
  - Uses the lx library provided by Uber for dependency injection of services and handlers

