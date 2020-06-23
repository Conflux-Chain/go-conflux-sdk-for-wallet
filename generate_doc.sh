godocdown -o tmp.md

sed -i "" 's/## Usage//g' tmp.md 
sed -i "" 's/#### type/### type/g' tmp.md


read -r -p "The api.md will be overwritten, are you sure ? [y/n] " input

case $input in
    [yY][eE][sS]|[yY])
		echo "Yes"
        mv ./tmp.md ./api.md
		;;

    [nN][oO]|[nN])
		echo "No"
       	;;

    *)
		echo "Invalid input..."
        rm ./tmp.md
		exit 1
		;;
esac