# SETA Training - API Documentation

## 🚀 Getting Started

### Base URLs
- **GraphQL**: `http://localhost:8080/graphql`
- **GraphQL Playground**: `http://localhost:8080/playground`
- **REST API**: `http://localhost:8080/api/v1`
- **Health Check**: `http://localhost:8080/health`

### Authentication
All protected endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

## 📊 GraphQL API (User Management)

### **Mutations**

#### Create User
```graphql
mutation {
  createUser(input: {
    username: "john_doe"
    email: "john@example.com"
    password: "securepassword123"
    role: manager  # or member
  }) {
    id
    username
    email
    role
    createdAt
    updatedAt
  }
}
```

#### Login
```graphql
mutation {
  login(input: {
    email: "john@example.com"
    password: "securepassword123"
  }) {
    user {
      id
      username
      email
      role
    }
    token
  }
}
```

#### Logout
```graphql
mutation {
  logout
}
```

### **Queries**

#### Fetch All Users
```graphql
query {
  fetchUsers {
    id
    username
    email
    role
    createdAt
    updatedAt
  }
}
```

#### Get Current User (Future)
```graphql
query {
  me {
    id
    username
    email
    role
  }
}
```

### **Types**

#### UserRole Enum
```graphql
enum UserRole {
  manager
  member
}
```

#### User Type
```graphql
type User {
  id: ID!
  username: String!
  email: String!
  role: UserRole!
  createdAt: String!
  updatedAt: String!
}
```

## 🔗 REST API (Team Management)

### **Authentication Required**
All team endpoints require authentication and appropriate permissions.

### **Endpoints**

#### Create Team
```http
POST /api/v1/teams
Authorization: Bearer <manager-token>
Content-Type: application/json

{
  "teamName": "Development Team",
  "managers": [
    {
      "managerId": "uuid-here",
      "managerName": "Manager Name"
    }
  ],
  "members": [
    {
      "memberId": "uuid-here",
      "memberName": "Member Name"
    }
  ]
}
```

**Response:**
```json
{
  "id": "team-uuid",
  "name": "Development Team",
  "created_at": "2025-07-24T10:16:52.057549Z",
  "updated_at": "2025-07-24T10:16:52.057549Z",
  "managers": [
    {
      "id": "user-uuid",
      "username": "manager1",
      "email": "manager@example.com",
      "role": "manager"
    }
  ],
  "members": []
}
```

#### Get Team
```http
GET /api/v1/teams/{teamId}
Authorization: Bearer <token>
```

#### Get All Teams
```http
GET /api/v1/teams
Authorization: Bearer <token>
```

#### Add Team Member
```http
POST /api/v1/teams/{teamId}/members
Authorization: Bearer <manager-token>
Content-Type: application/json

{
  "userId": "user-uuid"
}
```

#### Remove Team Member
```http
DELETE /api/v1/teams/{teamId}/members/{memberId}
Authorization: Bearer <manager-token>
```

#### Add Team Manager
```http
POST /api/v1/teams/{teamId}/managers
Authorization: Bearer <manager-token>
Content-Type: application/json

{
  "userId": "user-uuid"
}
```

#### Remove Team Manager
```http
DELETE /api/v1/teams/{teamId}/managers/{managerId}
Authorization: Bearer <manager-token>
```

## 🔒 Authorization Rules

### **User Roles**
- **Manager**: Can create teams, manage team members and managers
- **Member**: Can be added to teams, cannot manage teams

### **Permissions**
- **Team Creation**: Only managers can create teams
- **Team Management**: Only team managers can add/remove members and managers
- **Team Viewing**: All authenticated users can view teams

## 📝 Example Workflows

### **1. User Registration and Team Creation**

```bash
# 1. Create a manager user
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { createUser(input: { username: \"manager1\", email: \"manager@example.com\", password: \"password123\", role: manager }) { id username email role } }"
  }'

# 2. Login to get token
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { login(input: { email: \"manager@example.com\", password: \"password123\" }) { token } }"
  }'

# 3. Create a team
curl -X POST http://localhost:8080/api/v1/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token-from-step-2>" \
  -d '{
    "teamName": "Development Team",
    "managers": [],
    "members": []
  }'
```

### **2. Adding Members to Team**

```bash
# 1. Create a member user
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { createUser(input: { username: \"member1\", email: \"member@example.com\", password: \"password123\", role: member }) { id } }"
  }'

# 2. Add member to team (using manager token)
curl -X POST http://localhost:8080/api/v1/teams/{teamId}/members \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <manager-token>" \
  -d '{
    "userId": "<member-user-id>"
  }'
```

## ⚠️ Error Responses

### **Common Error Codes**
- `400 Bad Request`: Invalid input data
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

### **Error Response Format**
```json
{
  "error": "Error message description"
}
```

## 🔍 Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "healthy"
}
```

## 🚧 Future Endpoints (Planned)

### **Asset Management**
- `POST /api/v1/folders` - Create folder
- `GET /api/v1/folders/{id}` - Get folder
- `POST /api/v1/folders/{id}/notes` - Create note in folder
- `POST /api/v1/folders/{id}/share` - Share folder
- `GET /api/v1/teams/{id}/assets` - View team assets (managers only)

These endpoints will be implemented in future iterations following the same patterns established in the current codebase.
