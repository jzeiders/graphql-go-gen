// Example TypeScript file with GraphQL operations

import { gql } from '@apollo/client';

export const GET_USER_PROFILE = gql`
  query GetUserProfile($userId: ID!) {
    user(id: $userId) {
      id
      username
      email
      profile {
        bio
        avatarUrl
        location
      }
      posts {
        id
        title
        status
        createdAt
      }
    }
  }
`;

export const UPDATE_USER_BIO = gql`
  mutation UpdateUserBio($userId: ID!, $bio: String!) {
    updateUser(id: $userId, input: { profile: { bio: $bio } }) {
      id
      profile {
        bio
      }
    }
  }
`;

// GraphQL comment style
export const LIST_POSTS = /* GraphQL */ `
  query ListPosts($limit: Int = 10) {
    posts(query: null) {
      id
      title
      content
      author {
        id
        username
      }
      tags
      status
      createdAt
    }
  }
`;

// Fragment example
export const USER_FIELDS = gql`
  fragment UserFields on User {
    id
    username
    email
    createdAt
    updatedAt
  }
`;

export const GET_USER_WITH_FRAGMENT = gql`
  query GetUserWithFragment($id: ID!) {
    user(id: $id) {
      ...UserFields
      profile {
        bio
      }
    }
  }
`;