require "twitter"
require "json"
require "open-uri"
require "erb"
require "builder"

SITE_URL = "https://charlieegan3-twitter-rss.netlify.app/"

client = Twitter::REST::Client.new do |config|
  config.consumer_key        = ENV.fetch("TWITTER_ACCESS_API_KEY")
  config.consumer_secret     = ENV.fetch("TWITTER_ACCESS_API_SECRET_KEY")
  config.access_token        = ENV.fetch("TWITTER_ACCESS_TOKEN")
  config.access_token_secret = ENV.fetch("TWITTER_ACCESS_TOKEN_SECRET")
end


tweets = []

max_id = nil
10.times do
  opts = {count: 100, exclude_replies: true, include_rts: true}
  opts.merge!(max_id: max_id) if max_id

  t = client.home_timeline(opts)

  tweets.push(*t.map(&:to_h))

  max_id = t.last.id

  if (Time.now - t.last.created_at) / 60 / 60 / 24 > 3
    break
  end
end

# this is just here as it allows me to test regeneration without loading from
# the API again
File.write("tweets.json", tweets.to_json)

template = '
<p style="font-size: 0.8rem; font-family: sans-serif;">
  <img
    style="border-radius: 100rem; margin-bottom: -0.5rem; width: 1.7rem;"
    src="<%= tweet["user"]["profile_image_url_https"] %>">
  <%= tweet["user"]["name"] %>
  (@<%= tweet["user"]["screen_name"] %>)
  <a href="https://twitter.com/<%= tweet["user"]["screen_name"] %>/status/<%= tweet["id_str"] %>">
    <%= Time.parse(tweet["created_at"]).strftime("%H:%M") %>
  </a>
<p>

<blockquote style="font-family: sans-serif;">
  <p><%= tweet["text"] %></p>
</blockquote>
<hr/>
'

tweets = JSON.parse(File.read("tweets.json"))

days = tweets.group_by { |t| Time.parse(t["created_at"]).strftime("%Y-%m-%d") }
days.delete(Time.now.strftime("%Y-%m-%d")) # remove today

xml = Builder::XmlMarkup.new
xml.instruct! :xml, :version => '1.0'
out = xml.rss "version" => "2.0", "xmlns:atom" => "http://www.w3.org/2005/Atom" do
  xml.channel do
    xml.title "Twitter Timeline"
    xml.description "Automated rss feed from home timeline"
    xml.link SITE_URL
    xml.lastBuildDate Time.now.strftime("%a, %-d %b %Y %T %z")
    xml.tag!("atom:link", href: "#{SITE_URL}feed.xml", rel: "self", type: "application/rss+xml") {} # {} adds a closing tag

    days.each do |date, tweets|
      description = tweets.reverse.map do |t|
        b = binding
        b.local_variable_set(:tweet, t)

        renderer = ERB.new(template)

        renderer.result(b)
      end.join("\n")

      xml.item do
        xml.title date
        xml.description description
        xml.pubDate Time.parse(date).strftime("%a, %-d %b %Y %T %z")
        xml.link "#{SITE_URL}items/#{date}"
        xml.guid "#{SITE_URL}items/#{date}"
      end
    end
  end
end

puts out
