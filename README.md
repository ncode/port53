[![Go](https://github.com/ncode/port53/actions/workflows/go.yml/badge.svg)](https://github.com/ncode/port53/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ncode/port53)](https://goreportcard.com/report/github.com/ncode/port53)
[![codecov](https://codecov.io/gh/ncode/port53/branch/main/graph/badge.svg?token=S5Z0VTL3VY)](https://codecov.io/gh/ncode/port53)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)


# Port53
Dns management system

#### Why?

- I'm tired of building a different closed source dns service every other year
- Having a DNS api that handles multi-tenant environments
- Make possible to use either nsupdate or API on a user level to manage DNS entries
- I'd like to have a single DNS Api over different DNS servers

#### Roadmap

- [ ] Api
  - [x] Backend CRUD
  - [x] Domain CRUD
  - [x] Record CRUD
  - [ ] nsupdate
- [ ] DNS Interface
 - [ ] Validate queries against service
- [ ] User management API
  - [ ] User CRUD
    - [ ] nsuppdate Keys
    - [ ] API Keys
  - [ ] User permissions
- [ ] Agent
 - [ ] Backends
   - [ ] Bind
   - [ ] PowerDNS
   - [ ] NSD
 - [ ] Small/Medium size deployment architecture
 - [ ] Large scale deployment architecture
