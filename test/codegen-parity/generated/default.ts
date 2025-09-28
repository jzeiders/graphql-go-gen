export type GetUserQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type GetUserQuery = { __typename?: 'Query', user?: { __typename?: 'User', id: string, name: string, email: string, role: UserRole, posts: Array<{ __typename?: 'Post', id: string, title: string, published: boolean }>, profile?: { __typename?: 'Profile', bio?: string | null, avatar?: string | null } | null } | null };

export type GetUsersQueryVariables = Exact<{
  first?: InputMaybe<Scalars['Int']['input']>;
  after?: InputMaybe<Scalars['String']['input']>;
}>;


export type GetUsersQuery = { __typename?: 'Query', users: { __typename?: 'UserConnection', totalCount: number, edges: Array<{ __typename?: 'UserEdge', cursor: string, node: { __typename?: 'User', id: string, name: string, email: string, status: Status } }>, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, endCursor?: string | null } } };

export type SearchContentQueryVariables = Exact<{
  query: Scalars['String']['input'];
  limit?: InputMaybe<Scalars['Int']['input']>;
}>;


export type SearchContentQuery = { __typename?: 'Query', search: Array<
    | { __typename: 'User', id: string, name: string, email: string }
    | { __typename: 'Post', id: string, title: string, content: string, author: { __typename?: 'User', name: string } }
    | { __typename: 'Comment', id: string, content: string, author: { __typename?: 'User', name: string } }
  > };

export type CreateUserMutationVariables = Exact<{
  input: CreateUserInput;
}>;


export type CreateUserMutation = { __typename?: 'Mutation', createUser: { __typename?: 'User', id: string, name: string, email: string, role: UserRole, createdAt: any } };

export type UpdateUserMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateUserInput;
}>;


export type UpdateUserMutation = { __typename?: 'Mutation', updateUser?: { __typename?: 'User', id: string, name: string, email: string, status: Status, updatedAt: any } | null };

export type PublishPostMutationVariables = Exact<{
  postId: Scalars['ID']['input'];
}>;


export type PublishPostMutation = { __typename?: 'Mutation', publishPost?: { __typename?: 'Post', id: string, title: string, published: boolean, publishedAt?: any | null } | null };

export type OnUserCreatedSubscriptionVariables = Exact<{ [key: string]: never; }>;


export type OnUserCreatedSubscription = { __typename?: 'Subscription', userCreated: { __typename?: 'User', id: string, name: string, email: string, role: UserRole } };

export type OnCommentAddedSubscriptionVariables = Exact<{
  postId: Scalars['ID']['input'];
}>;


export type OnCommentAddedSubscription = { __typename?: 'Subscription', commentAdded: { __typename?: 'Comment', id: string, content: string, createdAt: any, author: { __typename?: 'User', name: string } } };

export type GetPostWithFragmentsQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type GetPostWithFragmentsQuery = { __typename?: 'Query', post?: { __typename?: 'Post', id: string, title: string, content: string, published: boolean, tags: Array<string>, author: { __typename?: 'User', id: string, name: string, email: string, role: UserRole, status: Status }, comments: Array<{ __typename?: 'Comment', id: string, content: string, author: { __typename?: 'User', id: string, name: string, email: string, role: UserRole, status: Status } }> } | null };

export type UserFieldsFragment = { __typename?: 'User', id: string, name: string, email: string, role: UserRole, status: Status };

export type PostFieldsFragment = { __typename?: 'Post', id: string, title: string, content: string, published: boolean, tags: Array<string>, author: { __typename?: 'User', id: string, name: string, email: string, role: UserRole, status: Status } };