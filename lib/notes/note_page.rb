require_relative "page"

class Notes::NotePage < Notes::Page
  attr_reader(
    :uid,
    :short_uid,
    :slug,
    :tags,
    :published_at,
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
    @published_at ||= attributes[:published_at]
    @title ||= attributes[:title]
    @body ||= attributes[:body]
    @url ||= attributes[:url]
  end
end
