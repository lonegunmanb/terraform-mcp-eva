package prompt

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func AddSolveAvmIssuePrompt(s *mcp.Server) {
	s.AddPrompt(&mcp.Prompt{
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "issue_number",
				Description: "The issue number to be processed, for example: `1234`. If not provided, the prompt will try to infer it from the context or ask the user to provide it.",
			},
			{
				Name:        "category",
				Description: "The category of the issue, for example: `bug`, `feature`, `doc`, 'chore', etc. If not provided, the prompt will try to infer it from the context or ask the user to provide it.",
			},
		},
		Description: "If you're processing changes to repo that contains `terraform` and `avm` in repo's name, and there is `avm` or `avm.bat` file in the root directory, you should use this prompt to get instructions on how to process the changes. The prompt will return a list of instructions that you can follow to process the changes.",
		Name:        "solve_avm_issue",
	}, func(ctx context.Context, session *mcp.ServerSession, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
		issueNumber := params.Arguments["issue_number"]
		category := params.Arguments["category"]
		return &mcp.GetPromptResult{
			Meta:        nil,
			Description: "",
			Messages: []*mcp.PromptMessage{
				&mcp.PromptMessage{
					Content: &mcp.TextContent{
						Text: fmt.Sprintf(`As an AVM development expert, you must strictly follow these steps:
Analyze the user's request: and extract the issue number from it.
The issue number is %s, and the category is %s.
Use git checkout -b <category>/<issue-number> to create and switch to a new branch.
Make all necessary code changes to resolve the issue.
If you are about to create new example under 'examples' directory, please ask for permission first. Don't forget to add '_footer.md' and '_header.md' files like other examples.'
[CRITICAL STEP] After all changes are complete, you must execute:
1. 'mapotf transform --mptf-dir git::https://github.com/lonegunmanb/common-mapotf-fix-for-terraform-vibe-coding.git --tf-dir .'
2. 'mapotf clean-backup --tf-dir .'
2. ./avm pre-commit (or './avm.bat pre-commit' if you on Windows').
3. the following sub-checks: ['tfvalidatecheck', 'lint'] with './avm ' or './avm.bat
If checks succeeds too then you can propose creating a Pull Request (PR). If it fails, report the failure message, try to solve the issues with best effort.
Now, please begin execution.`, issueNumber, category),
					},
					Role: "user",
				},
			},
		}, nil
	})
}
