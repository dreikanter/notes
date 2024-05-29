class Notes::Page
  attr_reader(
    :template,
    :layout,
    :local_path,
    :public_path,
    :canonical_path,
    :published_at,
    :title
  )

  def initialize(attributes = {})
    @template = attributes[:template]
    @layout = attributes[:layout]
    @local_path = attributes[:local_path]
    @public_path = attributes[:public_path]
    @canonical_path = attributes[:canonical_path]
    @attachments = attributes[:attachments]
    @published_at = attributes[:published_at]
    @title = attributes[:title]
  end

  def attachments
    @attachments ||= []
  end
end
