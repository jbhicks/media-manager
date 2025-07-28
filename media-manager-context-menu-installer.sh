#!/bin/bash
# media-manager-context-menu-installer.sh
# Adds 'media-manager' to right-click menu for common file managers

set -e

MEDIA_MANAGER_PATH="$(which media-manager || echo /usr/local/bin/media-manager)"

echo "Detected media-manager at: $MEDIA_MANAGER_PATH"

OS="$(uname)"
DESKTOP_ENV="$(echo $XDG_CURRENT_DESKTOP | tr '[:upper:]' '[:lower:]')"

# --- Linux: Nautilus (GNOME) ---
if [[ "$OS" == "Linux" && "$DESKTOP_ENV" == *"gnome"* && "$(which nautilus)" != "" ]]; then
    echo "Installing Nautilus script..."
    SCRIPTS_DIR="$HOME/.local/share/nautilus/scripts"
    mkdir -p "$SCRIPTS_DIR"
    cat > "$SCRIPTS_DIR/Media Manager" <<EOF
#!/bin/bash
"$MEDIA_MANAGER_PATH" "\$@"
EOF
    chmod +x "$SCRIPTS_DIR/Media Manager"
    echo "Nautilus script installed at $SCRIPTS_DIR/Media Manager"
fi

# --- Linux: Thunar (XFCE) ---
if [[ "$OS" == "Linux" && "$DESKTOP_ENV" == *"xfce"* && "$(which thunar)" != "" ]]; then
    echo "Configuring Thunar custom action..."
    # Thunar custom actions are stored in ~/.config/Thunar/uca.xml
    UCA_FILE="$HOME/.config/Thunar/uca.xml"
    if [[ -f "$UCA_FILE" ]]; then
        # Insert custom action if not present
        if ! grep -q "media-manager" "$UCA_FILE"; then
            sed -i '/<\/actions>/i \
    <action>\
      <icon>application-x-executable</icon>\
      <name>Open with Media Manager</name>\
      <command>'"$MEDIA_MANAGER_PATH"' %F</command>\
      <description>Open selected files/folders in Media Manager</description>\
      <patterns>*</patterns>\
      <startup-notify/>\
      <directories/>\
      <audio-files/>\
      <image-files/>\
      <video-files/>\
      <text-files/>\
      <other-files/>\
    </action>' "$UCA_FILE"
            echo "Thunar custom action added. You may need to restart Thunar."
        else
            echo "Thunar custom action already present."
        fi
    else
        echo "Thunar config file not found. Please create a custom action manually."
    fi
fi

# --- Linux: Dolphin (KDE) ---
if [[ "$OS" == "Linux" && "$DESKTOP_ENV" == *"kde"* && "$(which dolphin)" != "" ]]; then
    echo "Installing Dolphin service menu..."
    SERVICE_MENU_DIR="$HOME/.local/share/kservices5/ServiceMenus"
    mkdir -p "$SERVICE_MENU_DIR"
    cat > "$SERVICE_MENU_DIR/media-manager.desktop" <<EOF
[Desktop Entry]
Type=Service
ServiceTypes=KonqPopupMenu/Plugin
MimeType=inode/directory;application/octet-stream;video/*;image/*;
Actions=OpenWithMediaManager

[Desktop Action OpenWithMediaManager]
Name=Open with Media Manager
Exec=$MEDIA_MANAGER_PATH %F
Icon=application-x-executable
EOF
    echo "Dolphin service menu installed at $SERVICE_MENU_DIR/media-manager.desktop"
fi

# --- Windows: Explorer ---
if [[ "$OS" == "MINGW"* || "$OS" == "MSYS"* || "$OS" == "CYGWIN"* ]]; then
    echo "Generating Windows registry file for context menu..."
    REG_FILE="media-manager-context-menu.reg"
    cat > "$REG_FILE" <<EOF
Windows Registry Editor Version 5.00

[HKEY_CLASSES_ROOT\Directory\shell\MediaManager]
@="Open with Media Manager"
"Icon"="C:\\Path\\To\\media-manager.exe"

[HKEY_CLASSES_ROOT\Directory\shell\MediaManager\command]
@="\"C:\\Path\\To\\media-manager.exe\" \"%1\""
EOF
    echo "Registry file generated: $REG_FILE"
    echo "Edit the path to your media-manager.exe, then double-click the .reg file to install."
fi

# --- macOS: Finder ---
if [[ "$OS" == "Darwin" ]]; then
    echo "macOS detected."
    echo "Automated Finder integration is not supported in this script."
    echo "To add a Finder Quick Action:"
    echo "1. Open Automator, create a new 'Quick Action'."
    echo "2. Set 'Workflow receives current' to 'files or folders' in Finder."
    echo "3. Add a 'Run Shell Script' action with:"
    echo "   $MEDIA_MANAGER_PATH \"\$@\""
    echo "4. Save as 'Open with Media Manager'."
fi

echo "Context menu integration attempted for your OS/file manager."
