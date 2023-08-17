class Notes::Configuration
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

    def local_images_path
      expand_path("local_images_path")
    end

    def images_index_path
      File.join(local_images_path, "index.json")
    end

    private

    def expand_path(key)
      File.expand_path(fetch(key), File.dirname(file_name))
    end

    def fetch(key)
      data.fetch(key)
    end

    def data
      @data ||= begin
        abort "configuration file not found at #{file_name}" unless File.exist?(file_name)
        YAML.safe_load(File.read(file_name))
      end
    end

    def file_name
      @file_name ||= explicit_configuration_path || File.join(Dir.pwd, "configuration.yml")
    end

    def explicit_configuration_path
      ENV["NOTES_CONFIGURATION_PATH"]&.then { File.expand_path(_1) }
    end
  end
end
