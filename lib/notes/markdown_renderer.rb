class Notes::MarkdownRenderer < Redcarpet::Render::HTML
  include Rouge::Plugins::Redcarpet

  attr_reader :process_image

  def initialize(process_image:, extensions: {})
    super(extensions)
    @process_image = process_image
  end

  def image(link, title, alt_text)
    image_path = File.join(Notes::Configuration.site_root_path, process_image.call(link))
    "<img src=\"#{image_path}\" alt=\"#{alt_text}\" title=\"#{title}\">"
  end
end
