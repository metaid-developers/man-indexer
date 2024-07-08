FILE=/man/manindex-linux.tar.gz
if test -f "$FILE"; then
    echo "$FILE exist"
else
    wget --no-check-certificate $1
    tar -zxvf manindex-linux.tar.gz
    cp releases/linux/config.toml ./config.toml
fi