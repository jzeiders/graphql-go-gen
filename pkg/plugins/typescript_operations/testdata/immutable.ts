export type GetUserQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type GetUserQuery = { readonly __typename?: 'Query', readonly user?: { readonly __typename?: 'User', readonly id: string, readonly name: string, readonly email: string, readonly role: UserRole, readonly posts: ReadonlyArray<{ readonly __typename?: 'Post', readonly id: string, readonly title: string, readonly published: boolean }>, readonly profile?: { readonly __typename?: 'Profile', readonly bio?: string | null, readonly avatar?: string | null } | null } | null };

export type GetUsersQueryVariables = Exact<{
  first?: InputMaybe<Scalars['Int']['input']>;
  after?: InputMaybe<Scalars['String']['input']>;
}>;


export type GetUsersQuery = { readonly __typename?: 'Query', readonly users: { readonly __typename?: 'UserConnection', readonly totalCount: number, readonly edges: ReadonlyArray<{ readonly __typename?: 'UserEdge', readonly cursor: string, readonly node: { readonly __typename?: 'User', readonly id: string, readonly name: string, readonly email: string, readonly status: Status } }>, readonly pageInfo: { readonly __typename?: 'PageInfo', readonly hasNextPage: boolean, readonly endCursor?: string | null } } };

export type SearchContentQueryVariables = Exact<{
  query: Scalars['String']['input'];
  limit?: InputMaybe<Scalars['Int']['input']>;
}>;


export type SearchContentQuery = { readonly __typename?: 'Query', readonly search: ReadonlyArray<
    | { readonly __typename: 'User', readonly id: string, readonly name: string, readonly email: string }
    | { readonly __typename: 'Post', readonly id: string, readonly title: string, readonly content: string, readonly author: { readonly __typename?: 'User', readonly name: string } }
    | { readonly __typename: 'Comment', readonly id: string, readonly content: string, readonly author: { readonly __typename?: 'User', readonly name: string } }
  > };

export type CreateUserMutationVariables = Exact<{
  input: CreateUserInput;
}>;


export type CreateUserMutation = { readonly __typename?: 'Mutation', readonly createUser: { readonly __typename?: 'User', readonly id: string, readonly name: string, readonly email: string, readonly role: UserRole, readonly createdAt: any } };

export type UpdateUserMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateUserInput;
}>;


export type UpdateUserMutation = { readonly __typename?: 'Mutation', readonly updateUser?: { readonly __typename?: 'User', readonly id: string, readonly name: string, readonly email: string, readonly status: Status, readonly updatedAt: any } | null };

export type PublishPostMutationVariables = Exact<{
  postId: Scalars['ID']['input'];
}>;


export type PublishPostMutation = { readonly __typename?: 'Mutation', readonly publishPost?: { readonly __typename?: 'Post', readonly id: string, readonly title: string, readonly published: boolean, readonly publishedAt?: any | null } | null };

export type OnUserCreatedSubscriptionVariables = Exact<{ [key: string]: never; }>;


export type OnUserCreatedSubscription = { readonly __typename?: 'Subscription', readonly userCreated: { readonly __typename?: 'User', readonly id: string, readonly name: string, readonly email: string, readonly role: UserRole } };

export type OnCommentAddedSubscriptionVariables = Exact<{
  postId: Scalars['ID']['input'];
}>;


export type OnCommentAddedSubscription = { readonly __typename?: 'Subscription', readonly commentAdded: { readonly __typename?: 'Comment', readonly id: string, readonly content: string, readonly createdAt: any, readonly author: { readonly __typename?: 'User', readonly name: string } } };

export type UserFieldsFragment = { readonly __typename?: 'User', readonly id: string, readonly name: string, readonly email: string, readonly role: UserRole, readonly status: Status };

export type PostFieldsFragment = { readonly __typename?: 'Post', readonly id: string, readonly title: string, readonly content: string, readonly published: boolean, readonly tags: ReadonlyArray<string>, readonly author: { readonly __typename?: 'User', readonly id: string, readonly name: string, readonly email: string, readonly role: UserRole, readonly status: Status } };

export type GetPostWithFragmentsQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type GetPostWithFragmentsQuery = { readonly __typename?: 'Query', readonly post?: { readonly __typename?: 'Post', readonly id: string, readonly title: string, readonly content: string, readonly published: boolean, readonly tags: ReadonlyArray<string>, readonly comments: ReadonlyArray<{ readonly __typename?: 'Comment', readonly id: string, readonly content: string, readonly author: { readonly __typename?: 'User', readonly id: string, readonly name: string, readonly email: string, readonly role: UserRole, readonly status: Status } }>, readonly author: { readonly __typename?: 'User', readonly id: string, readonly name: string, readonly email: string, readonly role: UserRole, readonly status: Status } } | null };
