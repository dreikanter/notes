require "date"
require "fileutils"
require "http"
require "logger"
require "pry"
require "redcarpet"
require "rouge"
require "rouge/plugins/redcarpet"
require "securerandom"
require "tilt"
require "uri"
require "yaml"

$:.unshift(File.expand_path(__dir__))

module Notes
  autoload :CleanshotDownloader, "notes/cleanshot_downloader"
  autoload :Configuration, "notes/configuration"
  autoload :ImagesCache, "notes/images_cache"
  autoload :Logging, "notes/logging"
  autoload :MarkdownRenderer, "notes/markdown_renderer"
  autoload :NotePage, "notes/note_page"
  autoload :NotePageBuilder, "notes/note_page_builder"
  autoload :Page, "notes/page"
  autoload :RedirectPage, "notes/redirect_page"
  autoload :Site, "notes/site"
  autoload :SiteBuilder, "notes/site_builder"
  autoload :TagPage, "notes/tag_page"
end
