class Notes::NotePageBuilder
  attr_reader :source_file

  def initialize(source_file)
    @source_file = source_file
  end

  def build
    return unless public?

    Notes::NotePage.new(
      uid: uid,
      short_uid: short_uid,
      slug: slug,
      tags: tags,
      published_at: published_at,
      title: title,
      body: body,
      url: url,
      local_path: local_path,
      public_path: public_path,
      attachments: attachments
    )
  end

  private

  def public?
    metadata["public"]
  end

  def uid
    @uid ||= File.basename(source_file, ".*").match(/(^\d+_\d+)/).captures.first
  end

  def short_uid
    uid.gsub(/^\d+_/, "")
  end

  def slug
    @slug ||= slugify(metadata["slug"] || metadata["title"])
  end

  def slugify(string)
    string.to_s.strip.downcase.gsub(/[^a-z0-9\s]+/i, " ").strip.gsub(/\s+/, "-")
  end

  def tags
    metadata["tags"]&.uniq || []
  end

  def body
    Redcarpet::Markdown.new(
      markdown_renderer,
      fenced_code_blocks: true,
      autolink: true,
      strikethrough: true,
      space_after_headers: true,
      highlight: true
    ).render(source_content.gsub(FRONTMATTER_PATETRN, ""))
  end

  def markdown_renderer
    Notes::MarkdownRenderer.new(
      extensions: {with_toc_data: true},
      process_image: process_image
    )
  end

  def process_image
    lambda do |url|
      Notes::ImagesCache.new.get(url: url, page_uid: uid).tap { attachments << _1 }
    end
  end

  def attachments
    @attachments ||= []
  end

  def local_path
    @local_path ||= File.join([uid, slug, "index.html"].compact)
  end

  def public_path
    @public_path ||= File.join(site_root_path, local_path.gsub(/\/index.html$/, ""))
  end

  def url
    File.join(site_root_url, public_path)
  end

  def title
    (metadata["title"] || uid).gsub("`", "")
  end

  def published_at
    return if uid.empty?
    year, month, day = uid.match(/^(\d+)(\d\d)(\d\d)_/).captures
    Date.new(Integer(year), Integer(month.gsub(/^0+/, "")), Integer(day.gsub(/^0+/, ""))).to_time
  rescue StandardError => e
    raise "error parsing page timestamp: '#{source_file}'; error: #{e}"
  end

  def metadata
    @metadata ||= frontmatter? ? YAML.safe_load(source_content) : {}
  end

  def frontmatter?
    source_content.match?(FRONTMATTER_PATETRN)
  end

  def source_content
    @source_content ||= File.read(source_file)
  end

  FRONTMATTER_PATETRN = /\A(---\s*\n.*?\n?)^((---|\.\.\.)\s*$\n?)/m

  private_constant :FRONTMATTER_PATETRN

  def site_root_path
    Notes::Configuration.site_root_path
  end

  def site_root_url
    Notes::Configuration.site_root_url
  end
end
