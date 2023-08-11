class MarkdownRenderer < Redcarpet::Render::HTML
  include Rouge::Plugins::Redcarpet

  # def image(link, title, alt_text)
  #   new_link = ""
  #   super(new_link, title, alt_text)
  # end
end
