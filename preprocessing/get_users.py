import requests


def is_documentation_related(bio, keywords):
    if not bio:
        return False

    bio_lower = bio.lower()
    return any(keyword.lower() in bio_lower for keyword in keywords)


def get_contributors_with_bios(owner, repo, github_token, max_commits=12000):
    query_template = """
    {{
        repository(owner: "{owner}", name: "{repo}") {{
            defaultBranchRef {{
                target {{
                    ... on Commit {{
                        history(first: 100, after: {after_cursor}) {{
                            pageInfo {{
                                hasNextPage
                                endCursor
                            }}
                            nodes {{
                                author {{
                                    user {{
                                        login
                                        name
                                        bio
                                    }}
                                }}
                            }}
                        }}
                    }}
                }}
            }}
        }}
    }}
    """
    
   

    headers = {
        "Authorization": f"Bearer {github_token}",
        "Content-Type": "application/json",
    }

    graphql_url = "https://api.github.com/graphql"
    
    contributors = []
    commits_processed = 0
    has_next_page = True
    after_cursor = "null"

    while has_next_page and commits_processed < max_commits:
        query = query_template.format(owner=owner, repo=repo, after_cursor=after_cursor)
        graphql_payload = {"query": query}
        graphql_response = requests.post(graphql_url, headers=headers, json=graphql_payload)


        if graphql_response.status_code != 200:
            raise Exception(f"Error fetching data: {graphql_response.text}")

        response_data = graphql_response.json()
        if 'data' not in response_data:
            raise Exception(f"Error fetching data: {response_data}")

        commit_nodes = response_data["data"]["repository"]["defaultBranchRef"]["target"]["history"]["nodes"]

        for node in commit_nodes:
            author = node["author"]["user"]
            if author and author not in contributors:
                contributors.append(author)
        
        commits_processed += len(commit_nodes)

        page_info = response_data["data"]["repository"]["defaultBranchRef"]["target"]["history"]["pageInfo"]
        has_next_page = page_info["hasNextPage"]
        after_cursor = f'"{page_info["endCursor"]}"' if has_next_page else "null"

    return contributors



if __name__ == "__main__":
    owner = "MicrosoftDocs"
    #repo ="architecture-center"
    repo = "azure-devops-docs"
    # repo = "Virtualization-Documentation"
    # repo = "PowerShell-Docs"

    documentation_keywords = [
        "Writer", "Documentation", "Content Developer", "Documentation",
        "Information Curator", "Editor", "Knowledge Manager", "Content", "User Assistance Designer"
    ]

    contributors = get_contributors_with_bios(owner, repo, github_token)
    documentation_contributors = [contributor for contributor in contributors if is_documentation_related(contributor["bio"], documentation_keywords)]

    print(f"########## Code Contributors ##########")
    for contributor in contributors:
        print(f"Name: {contributor['name']}, Login: {contributor['login']}, Job Title: {contributor['bio']}")

    print(f"########## Writing to file ##########")
    with open("contributors.txt", "w") as f:
        for contributor in documentation_contributors:
            f.write(f"{contributor['name']}\n")
            f.write(f"{contributor['login']}\n")
