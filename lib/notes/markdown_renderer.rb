class Notes::MarkdownRenderer < Redcarpet::Render::HTML
  include Rouge::Plugins::Redcarpet

  def image(link, title, alt_text)
    "<img src=\"#{cached_image_url(link)}\" alt=\"#{alt_text}\" title=\"#{title}\">"
  end

  private

  def cached_image_url(link)
    Notes::ImagesCache.new.get(link)
  end
end
