export type GetUserVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type GetUser = { __typename?: 'Query', user?: { __typename?: 'User', id: string, name: string, email: string, role: UserRole, posts: Array<{ __typename?: 'Post', id: string, title: string, published: boolean }>, profile?: { __typename?: 'Profile', bio?: string | null, avatar?: string | null } | null } | null };

export type GetUsersVariables = Exact<{
  first?: InputMaybe<Scalars['Int']['input']>;
  after?: InputMaybe<Scalars['String']['input']>;
}>;


export type GetUsers = { __typename?: 'Query', users: { __typename?: 'UserConnection', totalCount: number, edges: Array<{ __typename?: 'UserEdge', cursor: string, node: { __typename?: 'User', id: string, name: string, email: string, status: Status } }>, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, endCursor?: string | null } } };

export type SearchContentVariables = Exact<{
  query: Scalars['String']['input'];
  limit?: InputMaybe<Scalars['Int']['input']>;
}>;


export type SearchContent = { __typename?: 'Query', search: Array<
    | { __typename: 'User', id: string, name: string, email: string }
    | { __typename: 'Post', id: string, title: string, content: string, author: { __typename?: 'User', name: string } }
    | { __typename: 'Comment', id: string, content: string, author: { __typename?: 'User', name: string } }
  > };

export type CreateUserVariables = Exact<{
  input: CreateUserInput;
}>;


export type CreateUser = { __typename?: 'Mutation', createUser: { __typename?: 'User', id: string, name: string, email: string, role: UserRole, createdAt: any } };

export type UpdateUserVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateUserInput;
}>;


export type UpdateUser = { __typename?: 'Mutation', updateUser?: { __typename?: 'User', id: string, name: string, email: string, status: Status, updatedAt: any } | null };

export type PublishPostVariables = Exact<{
  postId: Scalars['ID']['input'];
}>;


export type PublishPost = { __typename?: 'Mutation', publishPost?: { __typename?: 'Post', id: string, title: string, published: boolean, publishedAt?: any | null } | null };

export type OnUserCreatedVariables = Exact<{ [key: string]: never; }>;


export type OnUserCreated = { __typename?: 'Subscription', userCreated: { __typename?: 'User', id: string, name: string, email: string, role: UserRole } };

export type OnCommentAddedVariables = Exact<{
  postId: Scalars['ID']['input'];
}>;


export type OnCommentAdded = { __typename?: 'Subscription', commentAdded: { __typename?: 'Comment', id: string, content: string, createdAt: any, author: { __typename?: 'User', name: string } } };

export type UserFields = { __typename?: 'User', id: string, name: string, email: string, role: UserRole, status: Status };

export type PostFields = { __typename?: 'Post', id: string, title: string, content: string, published: boolean, tags: Array<string>, author: { __typename?: 'User', id: string, name: string, email: string, role: UserRole, status: Status } };

export type GetPostWithFragmentsVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type GetPostWithFragments = { __typename?: 'Query', post?: { __typename?: 'Post', id: string, title: string, content: string, published: boolean, tags: Array<string>, comments: Array<{ __typename?: 'Comment', id: string, content: string, author: { __typename?: 'User', id: string, name: string, email: string, role: UserRole, status: Status } }>, author: { __typename?: 'User', id: string, name: string, email: string, role: UserRole, status: Status } } | null };
