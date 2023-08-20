class Notes::MarkdownRenderer < Redcarpet::Render::HTML
  include Rouge::Plugins::Redcarpet

  attr_reader :process_image

  def initialize(process_image:, extensions: {})
    super(extensions)
    @process_image = process_image
  end

  def image(link, title, alt_text)
    file_name = process_image.call(link).fetch("file_name")
    "<img src=\"#{file_name}\" alt=\"#{alt_text}\" title=\"#{title}\">"
  end
end
