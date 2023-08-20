class Notes::NotePage < Notes::Page
  attr_reader(
    :uid,
    :short_uid,
    :slug,
    :tags,
    :title,
    :body,
    :url
  )

  def initialize(attributes = {})
    super(attributes.merge(template: "note.html", layout: "layout.html"))
    @uid ||= attributes[:uid]
    @short_uid ||= attributes[:short_uid]
    @slug ||= attributes[:slug]
    @tags ||= attributes[:tags]
    @title ||= attributes[:title]
    @body ||= attributes[:body]
    @url ||= attributes[:url]
  end
end
