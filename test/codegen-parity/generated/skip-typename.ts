export type GetUserQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type GetUserQuery = { user?: { id: string, name: string, email: string, role: UserRole, posts: Array<{ id: string, title: string, published: boolean }>, profile?: { bio?: string | null, avatar?: string | null } | null } | null };

export type GetUsersQueryVariables = Exact<{
  first?: InputMaybe<Scalars['Int']['input']>;
  after?: InputMaybe<Scalars['String']['input']>;
}>;


export type GetUsersQuery = { users: { totalCount: number, edges: Array<{ cursor: string, node: { id: string, name: string, email: string, status: Status } }>, pageInfo: { hasNextPage: boolean, endCursor?: string | null } } };

export type SearchContentQueryVariables = Exact<{
  query: Scalars['String']['input'];
  limit?: InputMaybe<Scalars['Int']['input']>;
}>;


export type SearchContentQuery = { search: Array<
    | { __typename: 'User', id: string, name: string, email: string }
    | { __typename: 'Post', id: string, title: string, content: string, author: { name: string } }
    | { __typename: 'Comment', id: string, content: string, author: { name: string } }
  > };

export type CreateUserMutationVariables = Exact<{
  input: CreateUserInput;
}>;


export type CreateUserMutation = { createUser: { id: string, name: string, email: string, role: UserRole, createdAt: any } };

export type UpdateUserMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateUserInput;
}>;


export type UpdateUserMutation = { updateUser?: { id: string, name: string, email: string, status: Status, updatedAt: any } | null };

export type PublishPostMutationVariables = Exact<{
  postId: Scalars['ID']['input'];
}>;


export type PublishPostMutation = { publishPost?: { id: string, title: string, published: boolean, publishedAt?: any | null } | null };

export type OnUserCreatedSubscriptionVariables = Exact<{ [key: string]: never; }>;


export type OnUserCreatedSubscription = { userCreated: { id: string, name: string, email: string, role: UserRole } };

export type OnCommentAddedSubscriptionVariables = Exact<{
  postId: Scalars['ID']['input'];
}>;


export type OnCommentAddedSubscription = { commentAdded: { id: string, content: string, createdAt: any, author: { name: string } } };

export type GetPostWithFragmentsQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type GetPostWithFragmentsQuery = { post?: { id: string, title: string, content: string, published: boolean, tags: Array<string>, author: { id: string, name: string, email: string, role: UserRole, status: Status }, comments: Array<{ id: string, content: string, author: { id: string, name: string, email: string, role: UserRole, status: Status } }> } | null };

export type UserFieldsFragment = { id: string, name: string, email: string, role: UserRole, status: Status };

export type PostFieldsFragment = { id: string, title: string, content: string, published: boolean, tags: Array<string>, author: { id: string, name: string, email: string, role: UserRole, status: Status } };