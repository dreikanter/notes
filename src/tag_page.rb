require_relative "./basic_page"

class TagPage < BasicPage
  attr_reader :tag

  def initialize(tag)
    @tag = tag
  end

  def title
    "##{tag}"
  end

  def published_at
    nil
  end
end
