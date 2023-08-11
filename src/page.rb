require "date"
require "redcarpet"
require "rouge"
require "rouge/plugins/redcarpet"
require_relative "./basic_page"
require_relative "./markdown_renderer"

class Page < BasicPage
  attr_reader :source_file, :configuration

  def initialize(source_file:, configuration:)
    @source_file = source_file
    @configuration = configuration
  end

  def public?
    metadata["public"]
  end

  def uid
    @uid ||= File.basename(source_file, ".*").match(/(^\d+_\d+)/).captures.first
  end

  def short_uid
    @short_uid ||= uid.gsub(/^\d+_/, "")
  end

  def slug
    @slug ||= slugify(metadata["slug"] || metadata["title"])
  end

  def tags
    @tags ||= metadata["tags"]&.uniq || []
  end

  def published_at
    @published_at ||= parse_published_at
  end

  def body
    @body ||= redcarpet_parser.render(source_content.gsub(FRONTMATTER_PATETRN, ""))
  end

  def page_file
    @page_file ||= File.join([uid, slug, "index.html"].compact)
  end

  def redirect_file
    @redirect_file ||= File.join(uid, "index.html")
  end

  def public_path
    @public_path ||= File.join(configuration.site_root_path, uid, slug)
  end

  def title
    @title ||= (metadata["title"] || uid).gsub("`", "")
  end

  def hide_from_toc?
    metadata["hide_from_toc"]
  end

  private

  def parse_published_at
    return if uid.empty?
    year, month, day = uid.match(/^(\d+)(\d\d)(\d\d)_/).captures
    Date.new(Integer(year), Integer(month.gsub(/^0+/, "")), Integer(day.gsub(/^0+/, ""))).to_time
  rescue StandardError => e
    raise "error parsing page timestamp: '#{source_file}'; error: #{e}"
  end

  def metadata
    @metadata ||= frontmatter? ? YAML.safe_load(source_content) : {}
  end

  def slugify(string)
    string.to_s.strip.downcase.gsub(/[^a-z0-9\s]+/i, " ").strip.gsub(/\s+/, "-")
  end

  def frontmatter?
    source_content.match?(FRONTMATTER_PATETRN)
  end

  def source_content
    @source_content ||= File.read(source_file)
  end

  def redcarpet_parser
    Redcarpet::Markdown.new(
      MarkdownRenderer.new(with_toc_data: true),
      fenced_code_blocks: true,
      autolink: true,
      strikethrough: true,
      space_after_headers: true,
      highlight: true
    )
  end

  FRONTMATTER_PATETRN = /\A(---\s*\n.*?\n?)^((---|\.\.\.)\s*$\n?)/m

  private_constant :FRONTMATTER_PATETRN
end
