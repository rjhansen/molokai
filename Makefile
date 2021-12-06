SHELL := /bin/bash

molokai:
	go build

install: molokai
	@if [ -f "molokai.conf" ]; then\
		sudo cp molokai /usr/local/bin;\
		sudo cp molokai.service /lib/systemd/system/molokai.service;\
		sudo ln -s /lib/systemd/system/molokai.service /etc/systemd/system/multi-user.target.wants/;\
		sudo cp molokai.conf /etc/molokai.conf;\
		sudo chmod 755 /usr/local/bin/molokai;\
		sudo chmod 600 /etc/molokai.conf;\
		sudo chmod 777 /etc/systemd/system/multi-user.target.wants/molokai.service;\
		sudo chown root:root /etc/molokai.conf;\
		sudo chown root:root /usr/local/bin/molokai;\
		sudo chown root:root /etc/systemd/system/multi-user.target.wants/molokai.service;\
		sudo systemctl start molokai;\
	else\
		echo "Create a molokai.conf and try again.";\
	fi;

uninstall:
	if [ -f "/etc/systemd/system/multi-user.target.wants/molokai.service" ]; then\
    	sudo systemctl stop molokai;\
	fi
	sudo rm -f /usr/local/bin/molokai
	sudo rm -f /etc/systemd/system/multi-user.target.wants/molokai.service
	sudo rm -f /lib/systemd/system/molokai.service
	sudo rm -f /etc/molokai.conf

clean: uninstall
	rm -f molokai

default: molokai
