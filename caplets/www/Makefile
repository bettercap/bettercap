all: facebook

facebook:
	wget -U "Mozilla/5.0 (Windows NT 5.2; rv:2.0.1) Gecko/20100101 Firefox/4.0.1" -S -r www.facebook.com
	find www.facebook.com -name "*.html" -print0 | xargs -0 sed -i "s/https:\/\/www.facebook.com//g"

clean:
	rm -rf www.facebook.com

