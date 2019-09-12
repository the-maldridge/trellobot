GitHub TrelloBot
================

This is a very simple bot that you can add to your GitHub repositories
to enforce the attachment of a trello card to your pull requests.
This can be useful if you use Trello instead of GitHub's built in
KanBan board to track work.  It will create a status object that
clears only once a card has been "attatched" via a comment to a PR.
This system does not actually check that the trello card exists or
that the referential integrity of the attachment is intact, but these
are both issues that are beyond the scope of this project (doing this
would also break the ability to just paste the link).

To use the bot install it somewhere that you will be able to make
visible to GitHub.  The service expects to pull two different values
from the environment:

* GITHUB_WEBHOOK_SECRET - This value is used to validate that it is
  actually GitHub calling and must match the value set in GitHub.
* GITHUB_PERSONAL_TOKEN - This value must be a personal access token
  with the repo:status scope.
