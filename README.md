# seta-training
SETA golang/nodejs training
# 🏗 Training Exercise: User, Team & Asset Management

## 🎯 Objective

Build a microservices-based system to manage users, teams, and digital assets—avoiding a monolithic design:

- Users can have roles: **manager** or **member**.
- Managers can create teams, add/remove members or other managers.
- Users can manage and share digital assets (folders & notes) with access control.

---

## ⚙ Proposed Microservice Architecture

- ✅ **GraphQL service**: For user management: create user, login, logout, fetch users, assign roles.
- ✅ **REST API**: For team management & asset management (folders, notes, sharing).

---

## 🧩 Functional Requirements

### 🔹 Auth & User Management Service (GraphQL)

- Create user:
  - `userId` (auto-generated)
  - `username`
  - `email` (unique)
  - `role`: "manager" or "member"
- Authentication:
  - Login, logout (JWT or session-based)
- User listing & query:
  - `fetchUsers` to get list of users
- Role assignment:
  - Manager: can create teams, manage users in teams
  - Member: can only be added to teams, no team management

---

### 🔹 Team Management Service (REST)

- Managers can:
  - Create teams
  - Add/remove members
  - Add/remove other managers (only main manager can do this)

Each team:

- `teamId`
- `teamName`
- `managers` (list)
- `members` (list)

---

### 🔹 Asset Management & Sharing (REST API)

- **Folders**: owned by users, contain notes
- **Notes**: belong to folders, have content
- Users can:
  - Share folders or individual notes with other users (read or write access)
  - Revoke access at any time
- When sharing a folder → all notes inside are also shared

**Managers**:

- Can view (read-only) all assets their team members have or can access
- Cannot edit unless explicitly shared with write access

---

## 🔑 Key Rules & Permissions

- Only authenticated users can use APIs.
- Managers can only manage users within their own teams.
- Members cannot create/manage teams.
- Only asset owners can manage sharing.

---

## 🛠 API Endpoints

### 📌 GraphQL: User Management

| Query/Mutation                      | Description             |
| ----------------------------------- | ----------------------- |
| `createUser(username, email, role)` | Create a new user       |
| `login(email, password)`            | Login and receive token |
| `logout()`                          | Logout current user     |
| `fetchUsers()`                      | List all users          |

---

### 📌 REST API: Team Management

| Method | Path                                 | Description        |
| ------ | ------------------------------------ | ------------------ |
| POST   | /teams                               | Create a team      |
| POST   | /teams/{teamId}/members              | Add member to team |
| DELETE | /teams/{teamId}/members/{memberId}   | Remove member      |
| POST   | /teams/{teamId}/managers             | Add manager        |
| DELETE | /teams/{teamId}/managers/{managerId} | Remove manager     |

#### ✅ Create team – request body:

```json
{
  "teamName": "string",
  "managers": [{"managerId": "string", "managerName": "string"}],
  "members": [{"memberId": "string", "memberName": "string"}]
}
```

---

### 📌 REST API: Asset Management

#### 🔹 Folder Management

| Method | Path                | Description                    |
| ------ | ------------------- | ------------------------------ |
| POST   | /folders            | Create new folder              |
| GET    | /folders/\:folderId | Get folder details             |
| PUT    | /folders/\:folderId | Update folder (name, metadata) |
| DELETE | /folders/\:folderId | Delete folder and its notes    |

#### 🔹 Note Management

| Method | Path                      | Description               |
| ------ | ------------------------- | ------------------------- |
| POST   | /folders/\:folderId/notes | Create note inside folder |
| GET    | /notes/\:noteId           | View note                 |
| PUT    | /notes/\:noteId           | Update note               |
| DELETE | /notes/\:noteId           | Delete note               |

#### 🔹 Sharing API

| Method | Path                               | Description                         |
| ------ | ---------------------------------- | ----------------------------------- |
| POST   | /folders/\:folderId/share          | Share folder with user (read/write) |
| DELETE | /folders/\:folderId/share/\:userId | Revoke folder sharing               |
| POST   | /notes/\:noteId/share              | Share single note                   |
| DELETE | /notes/\:noteId/share/\:userId     | Revoke note sharing                 |

#### 🔹 Manager-only APIs

| Method | Path                   | Description                                         |
| ------ | ---------------------- | --------------------------------------------------- |
| GET    | /teams/\:teamId/assets | View all assets that team members own or can access |
| GET    | /users/\:userId/assets | View all assets owned by or shared with user        |

---

## 🧩 Database Design Suggestion (PostgreSQL)

- Users: `userId`, `username`, `email`, `role`, `passwordHash`
- Teams: `teamId`, `teamName`
- team\_members, team\_managers: mapping tables
- Folders: `folderId`, `name`, `ownerId`
- Notes: `noteId`, `title`, `body`, `folderId`, `ownerId`
- folder\_shares, note\_shares: `userId`, `access` ("read" or "write")

---

## ✅ Development Requirements

- **Authentication**
  - Use JWT for authentication.
  - Validate and decode tokens, including expiration checks and claims verification—not limited to extracting user ID.

- **Authorization (RBAC)**
  - Enforce role validation before permitting team creation or manager assignment.
  - Restrict actions based on defined user roles.

- **Error Handling**
  - Handle duplicate email registration attempts.
  - Validate roles and flag unauthorized operations.
  - Return appropriate HTTP status codes and error messages.

- **Data Modeling**
  - Define models for the following entities:
    - `User`
    - `Team`
    - `Folder`
    - `Note`

- **Technology Stack**
  - Use the Go framework (Gin + GORM).
  - Apply clean architecture and separation of concerns for maintainability.

- **Logging and Monitoring**
  - Implement centralized logging.
  - Suggested stack: Loki + Grafana + Promtail (alternatives acceptable if justified).
  - Ensure visibility into API performance and system 

---
