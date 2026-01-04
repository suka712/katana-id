# KatanaID - Requirements

## What is it?

A plug-and-play identity toolkit for developers. Auth, avatars, usernames, bot protection.

---

## Internal

### Auth
- [ ] Email verification - **Khiem**
- [ ] Forgot password - Note: build frontend as well **Khiem**
- [ ] Rate limiting frontend `LoginPage` - `SignupPage` - Note: client can send req only upon changing information - add 3 second debounce - **Anh**
- [ ] Rate limiting backend `/login` - `/signup` - Note: Limit password retry attempts - "Too many failed attempts - Please try again later" - **Anh**

### Interface
- [ ] Interface to send contact info via "Contact" - Rate limit as well

### Documentation
- [ ] Contribution guide + `/docs/contribution` endpoint

### Chores
- [ ] Find replacement for Hero Image bento

## Features

### 1. Username & avatar AI generation
Gemini API
- [ ] End point `/identity/username` - POST { prompt }
- [ ] End point `/identity/avatar` - POST { prompt }
- [ ] Dashboard card for username prompt & display
- [ ] Dashboard card for avatar prompt & display

### 2. Fraud detection

- [ ] End point --- - POST { prompt + email list }
- [ ] Dashboard card for email list upload

### 3. CAPTCHA
---
End of text