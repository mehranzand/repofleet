Output repofleet-demo.gif

Require echo
Require printf
Require sleep

Set Shell "bash"
Set FontSize 22
Set Width 1280
Set Height 720
Set Padding 32
Set TypingSpeed 60ms
Set Theme "Catppuccin Mocha"
Set Margin 0

Sleep 500ms

Type "rf"
Enter
Sleep 1s

Type "rf repo list"
Enter
Sleep 1.5s

Type "rf repo add /home/mehran/flivo -n vg-backend"
Enter
Sleep 1.5s

Type "rf repo add /home/mehran/repofleet -n rf"
Enter
Sleep 1.5s

Type "rf repo add /home/mehran/pulseup -n docker-monitoring"
Enter
Sleep 1.5s

Type "rf workspace config -h"
Enter
Sleep 1.5s

Type "rf workspace config --branch-pattern {type}/issue-{issue}-{description}"
Enter
Sleep 1.5s

Type "rf workspace config"
Enter
Sleep 1.5s

Type "rf issue -h"
Enter
Sleep 1.5s

Type "rf issue create 12345 --description 'auth token refresh' --type 'feat' --repo 'rf,vg-backend'"
Enter
Sleep 2s

Type "rf issue status"
Enter
Sleep 2s

Type "rf issue goto"
Enter
Sleep 2s

Type "rf issue switch"
Enter
Sleep 2s

Type "rf issue archive 12345"
Enter
Sleep 3s
