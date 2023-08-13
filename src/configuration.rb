require "uri"
require "yaml"

class Configuration
  class << self
    def notes_path
      expand_path("notes_path")
    end

    def build_path
      expand_path("build_path")
    end

    def templates_path
      expand_path("templates_path")
    end

    def site_root_url
      fetch("site_root_url").gsub(/\/$/, "")
    end

    def site_root_path
      URI.parse(fetch("site_root_url")).path.then { _1.empty? ? "/" : _1 }
    end

    def site_name
      fetch("site_name")
    end

    def author_name
      fetch("author_name")
    end

    def feed_url
      "#{site_root_url}#{feed_path}"
    end

    def feed_path
      File.join(site_root_path, "feed.xml")
    end

    private

    def expand_path(key)
      File.expand_path(fetch(key), root_path)
    end

    def fetch(key)
      data.fetch(key)
    end

    def data
      @data ||= YAML.safe_load(File.read(file_name))
    end

    def file_name
      File.join(root_path, "configuration.yml")
    end

    def root_path
      @root_path ||= ENV.fetch("NOTES_CONFIGURATION_PATH") { Dir.pwd }
    end
  end
end
