class Notes::TagPage < Notes::Page
  attr_reader :tag

  def initialize(attributes = {})
    super(attributes.merge(template: "tag.html", layout: "layout.html"))
    @tag = attributes.fetch(:tag)
  end
end
