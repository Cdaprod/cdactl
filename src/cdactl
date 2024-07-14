#!/bin/bash
# File: /usr/local/bin/cdactl

source /usr/local/lib/cda-common.sh

# Get the device architecture and hostname
DEVICE_ARCH=$(uname -m)
HOSTNAME=$(uname -n)
BRANCH_NAME="${HOSTNAME}/${DEVICE_ARCH}"

# Define the cfg function
cfg() {
    /usr/bin/git --git-dir=$HOME/.cfg/ --work-tree=$HOME "$@"
}

function show_cdactl_usage {
    echo -e "${BLUE}Usage: cdactl [command] [options]${NC}"
    echo -e "${YELLOW}Commands:${NC}"
    echo -e "  ${GREEN}1. network${NC}    - Manage network connections"
    echo -e "  ${GREEN}2. ssh${NC}        - SSH into devices"
    echo -e "  ${GREEN}3. update${NC}     - Update system packages"
    echo -e "  ${GREEN}4. backup${NC}     - Manage backups"
    echo -e "  ${GREEN}5. monitor${NC}    - Monitor system resources"
    echo -e "  ${GREEN}6. dotfiles${NC}   - Manage dotfiles (init, add, pull, sync)"
    echo -e "  ${GREEN}7. cred${NC}       - Manage credentials (store, retrieve)"
    echo -e "  ${GREEN}8. help${NC}       - Show this help message"
}

function set_branch_based_on_device {
    # Automatically switch to a branch named after the device architecture and hostname
    if [ ! -d "$HOME/.cfg" ]; then
        print_error "Dotfiles repository not initialized. Run 'cdactl dotfiles init' first."
        exit 1
    fi

    BRANCH_EXISTS=$(cfg branch --list | grep "$BRANCH_NAME")
    if [ -z "$BRANCH_EXISTS" ]; then
        print_header "Creating new branch for device: $BRANCH_NAME"
        cfg checkout -b "$BRANCH_NAME"
        cfg commit --allow-empty -m "Initial commit on $BRANCH_NAME"
        cfg push --set-upstream origin "$BRANCH_NAME"
    else
        print_header "Switching to branch for device: $BRANCH_NAME"
        cfg checkout "$BRANCH_NAME"
    fi
    check_command_status "Branch set to $BRANCH_NAME"
}

function network_command {
    case "$1" in
        status)
            print_header "Network Status"
            ip -c addr show
            ;;
        restart)
            print_header "Restarting Network"
            sudo systemctl restart NetworkManager
            check_command_status "Network restart"
            ;;
        *)
            print_error "Invalid network command. Use: status or restart"
            ;;
    esac
}

function ssh_command {
    print_header "Connecting to device: $1"
    ssh "$1"
}

function update_command {
    print_header "Updating System Packages"
    sudo apt update && sudo apt upgrade -y
    check_command_status "System update"
}

function backup_command {
    BACKUP_DIR="$HOME/backup"
    case "$1" in
        create)
            print_header "Creating Backup"
            mkdir -p "$BACKUP_DIR"
            tar --exclude='minio-persistent-data' -czvf "$BACKUP_DIR/backup_$(date +%Y%m%d).tar.gz" "$HOME"
            check_command_status "Backup creation"
            ;;
        restore)
            if [ -z "$2" ]; then
                print_error "No backup file specified for restore."
                exit 1
            fi
            if [ ! -f "$BACKUP_DIR/$2" ]; then
                print_error "Backup file not found: $2"
                exit 1
            fi
            print_header "Restoring from Backup"
            tar -xzvf "$BACKUP_DIR/$2" -C "$HOME"
            check_command_status "Backup restoration"
            ;;
        *)
            print_error "Invalid backup command. Use: create or restore"
            ;;
    esac
}

function monitor_command {
    print_header "System Resource Monitor"
    top -bn1 | head -n 20
}

function dotfiles_command {
    case "$1" in
        init)
            print_header "Initializing dotfiles repository"
            git init --bare $HOME/.cfg
            echo "alias cfg='/usr/bin/git --git-dir=$HOME/.cfg/ --work-tree=$HOME'" >> $HOME/.bashrc
            source $HOME/.bashrc
            cfg remote add origin https://github.com/Cdaprod/cda.cfg.git || print_warning "Remote origin already exists."
            check_command_status "Dotfiles initialization"
            ;;
        add)
            set_branch_based_on_device
            if [ -z "$2" ]; then
                print_error "No dotfile specified for addition."
                exit 1
            fi
            print_header "Adding dotfile: $2"
            cfg add "$2"
            cfg commit -m "Add $2"
            cfg push origin "$BRANCH_NAME"
            check_command_status "Dotfile added and pushed"
            ;;
        pull)
            set_branch_based_on_device
            print_header "Pulling dotfiles from remote"
            cfg pull origin "$BRANCH_NAME"
            check_command_status "Dotfiles pulled"
            ;;
        sync)
            set_branch_based_on_device
            print_header "Syncing dotfiles with remote"
            cfg add -A
            cfg commit -m "Sync dotfiles"
            cfg push origin "$BRANCH_NAME"
            check_command_status "Dotfiles synced"
            ;;
        *)
            print_error "Invalid dotfiles command. Use: init, add, pull, or sync"
            ;;
    esac
}

# Function to read password securely
read_password() {
    prompt=$1
    echo -n "$prompt: "
    stty -echo
    read password
    stty echo
    echo
    echo $password
}

function cred_command {
    case "$1" in
        store)
            SERVICE="$2"
            USERNAME="$3"
            if [ -z "$SERVICE" ] || [ -z "$USERNAME" ]; then
                print_error "Usage: cdactl cred store [service] [username]"
                exit 1
            fi
            PASSWORD=$(read_password "Input Secure Key")
            echo "url=https://$SERVICE" >> ~/.git-credentials
            echo "username=$USERNAME" >> ~/.git-credentials
            echo "password=$PASSWORD" >> ~/.git-credentials
            git config --global credential.helper store
            check_command_status "Credentials for $SERVICE stored successfully"
            ;;
        retrieve)
            print_header "Retrieving stored credentials"
            cat ~/.git-credentials
            ;;
        *)
            print_error "Invalid cred command. Use: store or retrieve"
            ;;
    esac
}

case "$1" in
    network)
        network_command "$2"
        ;;
    ssh)
        ssh_command "$2"
        ;;
    update)
        update_command
        ;;
    backup)
        backup_command "$2"
        ;;
    monitor)
        monitor_command
        ;;
    dotfiles)
        dotfiles_command "$2" "$3" "$4" "$5"
        ;;
    cred)
        cred_command "$2" "$3" "$4"
        ;;
    help|*)
        show_cdactl_usage
        ;;
esac