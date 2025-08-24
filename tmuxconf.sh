MY_SESSION=$(tmux list-sessions | grep "poeacts")
if [[ ! $MY_SESSION ]]; then
		# create a new session and `-d`etach
		tmux new-session -d -s poeacts
fi
tmux attach-session -d -t poeacts
