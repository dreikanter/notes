require "date"
require "fileutils"
require "redcarpet"
require "rouge"
require "rouge/plugins/redcarpet"
require "tilt"
require "uri"
require "yaml"

module Notes
  Page = ::Data.define(
    :template,
    :layout,
    :local_path,
    :public_path
  )

  NotePage = ::Data.define(
    *Page.members,
    :uid,
    :short_uid,
    :slug,
    :tags,
    :published_at,
    :title,
    :body,
    :url
  )

  TagPage = ::Data.define(*Page.members, :tag)

  RedirectPage = ::Data.define(*Page.members, :redirect_to)

  require_relative "./notes/configuration"
  require_relative "./notes/markdown_parser"
  require_relative "./notes/markdown_renderer"
  require_relative "./notes/note_page_builder"
  require_relative "./notes/site"
  require_relative "./notes/site_builder"
end
