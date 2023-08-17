require_relative "page"

class Notes::RedirectPage < Notes::Page
  attr_reader :redirect_to

  def initialize(attributes = {})
    super(attributes.merge(template: "redirect.html", layout: nil))
    @redirect_to = attributes.fetch(:redirect_to)
  end
end
