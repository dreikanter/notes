require "redcarpet"
require "rouge"
require "rouge/plugins/redcarpet"

class MarkdownRenderer < Redcarpet::Render::HTML
  include Rouge::Plugins::Redcarpet

  def image(link, title, alt_text)
    "<img src=\"#{link}\" alt=\"#{alt_text}\" title=\"#{title}\">"
  end
end
