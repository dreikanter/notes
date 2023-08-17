class Notes::MarkdownRenderer < Redcarpet::Render::HTML
  include Rouge::Plugins::Redcarpet

  attr_reader :process_image

  def initialize(process_image:, extensions: {})
    super(extensions)
    @process_image = process_image
  end

  def image(link, title, alt_text)
    "<img src=\"#{process_image.call(link)}\" alt=\"#{alt_text}\" title=\"#{title}\">"
  end
end
