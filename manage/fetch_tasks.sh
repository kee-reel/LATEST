# !/bin/bash
if [[ -z "$(ls tests)" ]]; then
	cd tests # Go inside
	git clone $GIT_REPO $GIT_REPO_FOLDER # Clone sample project
	git clone $TEMPLATES_GIT_REPO templates # Clone templates
	cd ..
fi
